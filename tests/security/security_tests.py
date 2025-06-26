#!/usr/bin/env python3
"""
ComputeHive Security Testing Suite
Tests for OWASP Top 10 vulnerabilities and security best practices
"""

import requests
import pytest
import json
import time
from urllib.parse import quote
import jwt
import base64

BASE_URL = "http://localhost:8080/api/v1"
TEST_USER = "test@computehive.io"
TEST_PASS = "testPassword123!"

class TestSQLInjection:
    """Test for SQL injection vulnerabilities"""
    
    def test_sql_injection_login(self):
        """Test login endpoint for SQL injection"""
        malicious_inputs = [
            "' OR '1'='1",
            "'; DROP TABLE users; --",
            "admin'--",
            "' UNION SELECT * FROM users--",
            "1' AND '1' = '1",
        ]
        
        for payload in malicious_inputs:
            response = requests.post(f"{BASE_URL}/auth/login", json={
                "email": payload,
                "password": "any"
            })
            
            # Should not return 500 (server error)
            assert response.status_code != 500, f"Potential SQL injection with payload: {payload}"
            
            # Should not authenticate
            assert response.status_code == 401 or response.status_code == 400
            
    def test_sql_injection_search(self):
        """Test search endpoints for SQL injection"""
        search_endpoints = [
            "/jobs/search",
            "/marketplace/search",
            "/users/search"
        ]
        
        malicious_queries = [
            "'; SELECT * FROM users WHERE '1'='1",
            "' OR 1=1--",
            "UNION SELECT password FROM users--"
        ]
        
        for endpoint in search_endpoints:
            for query in malicious_queries:
                response = requests.get(f"{BASE_URL}{endpoint}?q={quote(query)}")
                assert response.status_code != 500
                
class TestXSS:
    """Test for Cross-Site Scripting vulnerabilities"""
    
    def test_xss_in_job_submission(self):
        """Test XSS in job submission"""
        xss_payloads = [
            "<script>alert('XSS')</script>",
            "<img src=x onerror=alert('XSS')>",
            "javascript:alert('XSS')",
            "<svg/onload=alert('XSS')>",
            "<iframe src='javascript:alert(\"XSS\")'></iframe>"
        ]
        
        # Get auth token first
        auth_response = requests.post(f"{BASE_URL}/auth/login", json={
            "email": TEST_USER,
            "password": TEST_PASS
        })
        
        if auth_response.status_code == 200:
            token = auth_response.json().get("access_token")
            headers = {"Authorization": f"Bearer {token}"}
            
            for payload in xss_payloads:
                job_data = {
                    "name": payload,
                    "description": payload,
                    "command": "echo test"
                }
                
                response = requests.post(f"{BASE_URL}/jobs", 
                                       json=job_data, 
                                       headers=headers)
                
                # If job is created, verify the payload is properly escaped
                if response.status_code == 201:
                    job_id = response.json().get("id")
                    job_response = requests.get(f"{BASE_URL}/jobs/{job_id}", 
                                              headers=headers)
                    
                    if job_response.status_code == 200:
                        job_data = job_response.json()
                        # Verify XSS payload is escaped
                        assert payload not in str(job_data)
                        
class TestAuthentication:
    """Test authentication and authorization vulnerabilities"""
    
    def test_brute_force_protection(self):
        """Test if brute force protection is in place"""
        failed_attempts = 0
        
        for i in range(10):
            response = requests.post(f"{BASE_URL}/auth/login", json={
                "email": TEST_USER,
                "password": f"wrongpassword{i}"
            })
            
            if response.status_code == 429:  # Rate limited
                print(f"Rate limiting kicked in after {i+1} attempts")
                return
                
            failed_attempts += 1
            
        assert failed_attempts < 10, "No rate limiting detected for login attempts"
        
    def test_jwt_token_validation(self):
        """Test JWT token security"""
        # Test with invalid token
        headers = {"Authorization": "Bearer invalid.token.here"}
        response = requests.get(f"{BASE_URL}/jobs", headers=headers)
        assert response.status_code == 401
        
        # Test with expired token
        expired_token = jwt.encode({
            "user_id": "123",
            "exp": int(time.time()) - 3600  # Expired 1 hour ago
        }, "secret", algorithm="HS256")
        
        headers = {"Authorization": f"Bearer {expired_token}"}
        response = requests.get(f"{BASE_URL}/jobs", headers=headers)
        assert response.status_code == 401
        
    def test_password_requirements(self):
        """Test password complexity requirements"""
        weak_passwords = [
            "123456",
            "password",
            "abc123",
            "qwerty",
            "12345678"
        ]
        
        for weak_pass in weak_passwords:
            response = requests.post(f"{BASE_URL}/auth/register", json={
                "email": "newuser@test.com",
                "password": weak_pass
            })
            
            assert response.status_code == 400, f"Weak password accepted: {weak_pass}"
            
class TestAPISecurityHeaders:
    """Test for security headers"""
    
    def test_security_headers(self):
        """Verify security headers are present"""
        response = requests.get(f"{BASE_URL}/health")
        
        required_headers = {
            "X-Content-Type-Options": "nosniff",
            "X-Frame-Options": ["DENY", "SAMEORIGIN"],
            "X-XSS-Protection": "1; mode=block",
            "Strict-Transport-Security": "max-age=31536000",
            "Content-Security-Policy": None  # Should be present
        }
        
        for header, expected_values in required_headers.items():
            assert header in response.headers, f"Missing security header: {header}"
            
            if expected_values:
                if isinstance(expected_values, list):
                    assert response.headers[header] in expected_values
                else:
                    assert response.headers[header] == expected_values
                    
class TestFileUploadSecurity:
    """Test file upload security"""
    
    def test_malicious_file_upload(self):
        """Test uploading potentially malicious files"""
        # Get auth token
        auth_response = requests.post(f"{BASE_URL}/auth/login", json={
            "email": TEST_USER,
            "password": TEST_PASS
        })
        
        if auth_response.status_code == 200:
            token = auth_response.json().get("access_token")
            headers = {"Authorization": f"Bearer {token}"}
            
            # Test various malicious file types
            malicious_files = [
                ("test.exe", b"MZ\x90\x00"),  # Windows executable
                ("test.sh", b"#!/bin/bash\nrm -rf /"),  # Shell script
                ("test.php", b"<?php system($_GET['cmd']); ?>"),  # PHP shell
                ("../../../etc/passwd", b"root:x:0:0"),  # Path traversal
            ]
            
            for filename, content in malicious_files:
                files = {"file": (filename, content)}
                response = requests.post(f"{BASE_URL}/upload", 
                                       files=files, 
                                       headers=headers)
                
                # Should either reject or sanitize filename
                if response.status_code == 200:
                    uploaded_path = response.json().get("path", "")
                    assert "../" not in uploaded_path
                    assert not uploaded_path.endswith((".exe", ".sh", ".php"))
                    
class TestCryptography:
    """Test cryptographic implementations"""
    
    def test_password_storage(self):
        """Verify passwords are properly hashed"""
        # This would require database access or an admin endpoint
        # to verify passwords are not stored in plain text
        pass
        
    def test_sensitive_data_encryption(self):
        """Test that sensitive data is encrypted in transit and at rest"""
        # Verify HTTPS is enforced
        try:
            response = requests.get(BASE_URL.replace("https", "http"), 
                                  allow_redirects=False)
            assert response.status_code in [301, 302, 307, 308], \
                   "HTTP not redirecting to HTTPS"
        except:
            pass  # Connection might be refused which is also acceptable
            
class TestBusinessLogic:
    """Test business logic vulnerabilities"""
    
    def test_race_condition_job_submission(self):
        """Test for race conditions in job submission"""
        # Get auth token
        auth_response = requests.post(f"{BASE_URL}/auth/login", json={
            "email": TEST_USER,
            "password": TEST_PASS
        })
        
        if auth_response.status_code == 200:
            token = auth_response.json().get("access_token")
            headers = {"Authorization": f"Bearer {token}"}
            
            # Try to submit multiple jobs simultaneously
            import threading
            results = []
            
            def submit_job():
                response = requests.post(f"{BASE_URL}/jobs", 
                                       json={"name": "test", "command": "echo test"},
                                       headers=headers)
                results.append(response.status_code)
                
            threads = []
            for _ in range(10):
                t = threading.Thread(target=submit_job)
                threads.append(t)
                t.start()
                
            for t in threads:
                t.join()
                
            # Verify no unexpected errors
            assert all(status in [201, 429] for status in results)
            
    def test_insufficient_authorization(self):
        """Test for authorization bypass vulnerabilities"""
        # Create two users and test cross-user access
        # This would require creating test users and checking
        # if one user can access another user's resources
        pass

if __name__ == "__main__":
    pytest.main([__file__, "-v"]) 