#!/usr/bin/env python3
"""
ComputeHive Chaos Engineering Test Suite
Tests system resilience under various failure scenarios
"""

import time
import random
import subprocess
import requests
import threading
import pytest
from datetime import datetime
import psutil
import docker

class ChaosMonkey:
    """Base class for chaos engineering operations"""
    
    def __init__(self, target_services=None, chaos_level="light"):
        self.target_services = target_services or []
        self.chaos_level = chaos_level
        self.docker_client = docker.from_env()
        self.base_url = "http://localhost:8080/api/v1"
        
    def get_chaos_probability(self):
        """Get probability of chaos based on level"""
        levels = {
            "light": 0.05,
            "medium": 0.20,
            "heavy": 0.50,
            "extreme": 0.80
        }
        return levels.get(self.chaos_level, 0.05)
        
class NetworkChaos(ChaosMonkey):
    """Network-related chaos scenarios"""
    
    def simulate_packet_loss(self, interface="eth0", loss_percent=10):
        """Simulate packet loss on network interface"""
        print(f"Simulating {loss_percent}% packet loss on {interface}")
        cmd = f"tc qdisc add dev {interface} root netem loss {loss_percent}%"
        subprocess.run(cmd.split(), capture_output=True)
        
    def simulate_network_delay(self, interface="eth0", delay_ms=100):
        """Add network latency"""
        print(f"Adding {delay_ms}ms delay to {interface}")
        cmd = f"tc qdisc add dev {interface} root netem delay {delay_ms}ms"
        subprocess.run(cmd.split(), capture_output=True)
        
    def simulate_bandwidth_limit(self, interface="eth0", rate="1mbit"):
        """Limit network bandwidth"""
        print(f"Limiting bandwidth to {rate} on {interface}")
        cmd = f"tc qdisc add dev {interface} root tbf rate {rate} burst 32kbit latency 400ms"
        subprocess.run(cmd.split(), capture_output=True)
        
    def partition_network(self, service_groups):
        """Simulate network partition between service groups"""
        print(f"Creating network partition between {service_groups}")
        # This would use iptables rules to block traffic between groups
        for group1 in service_groups:
            for group2 in service_groups:
                if group1 != group2:
                    # Block traffic between groups
                    cmd = f"iptables -A INPUT -s {group1} -d {group2} -j DROP"
                    subprocess.run(cmd.split(), capture_output=True)
                    
    def restore_network(self, interface="eth0"):
        """Restore normal network conditions"""
        print(f"Restoring network on {interface}")
        cmd = f"tc qdisc del dev {interface} root"
        subprocess.run(cmd.split(), capture_output=True)
        
class ServiceChaos(ChaosMonkey):
    """Service-level chaos scenarios"""
    
    def kill_random_service(self):
        """Kill a random service container"""
        containers = self.docker_client.containers.list()
        target_containers = [c for c in containers if any(
            service in c.name for service in self.target_services
        )]
        
        if target_containers and random.random() < self.get_chaos_probability():
            victim = random.choice(target_containers)
            print(f"Killing container: {victim.name}")
            victim.kill()
            return victim.name
        return None
        
    def pause_random_service(self, duration=30):
        """Pause a random service for specified duration"""
        containers = self.docker_client.containers.list()
        target_containers = [c for c in containers if any(
            service in c.name for service in self.target_services
        )]
        
        if target_containers and random.random() < self.get_chaos_probability():
            victim = random.choice(target_containers)
            print(f"Pausing container: {victim.name} for {duration}s")
            victim.pause()
            time.sleep(duration)
            victim.unpause()
            return victim.name
        return None
        
    def restart_random_service(self):
        """Restart a random service"""
        containers = self.docker_client.containers.list()
        target_containers = [c for c in containers if any(
            service in c.name for service in self.target_services
        )]
        
        if target_containers and random.random() < self.get_chaos_probability():
            victim = random.choice(target_containers)
            print(f"Restarting container: {victim.name}")
            victim.restart()
            return victim.name
        return None
        
class ResourceChaos(ChaosMonkey):
    """Resource-related chaos scenarios"""
    
    def consume_cpu(self, cores=1, duration=60):
        """Consume CPU resources"""
        print(f"Consuming {cores} CPU cores for {duration}s")
        
        def cpu_burn():
            end_time = time.time() + duration
            while time.time() < end_time:
                _ = sum(i*i for i in range(1000000))
                
        threads = []
        for _ in range(cores):
            t = threading.Thread(target=cpu_burn)
            t.start()
            threads.append(t)
            
        for t in threads:
            t.join()
            
    def consume_memory(self, size_mb=1000, duration=60):
        """Consume memory resources"""
        print(f"Consuming {size_mb}MB memory for {duration}s")
        
        # Allocate memory
        data = []
        chunk_size = 10 * 1024 * 1024  # 10MB chunks
        chunks_needed = size_mb // 10
        
        for _ in range(chunks_needed):
            data.append(bytearray(chunk_size))
            
        time.sleep(duration)
        del data  # Release memory
        
    def fill_disk(self, path="/tmp/chaos", size_mb=1000):
        """Fill disk space"""
        print(f"Writing {size_mb}MB to disk at {path}")
        
        with open(path, "wb") as f:
            for _ in range(size_mb):
                f.write(bytearray(1024 * 1024))  # Write 1MB at a time
                
    def simulate_disk_failure(self, device="/dev/sdb"):
        """Simulate disk I/O errors"""
        print(f"Simulating disk failure on {device}")
        # This would use device mapper to inject I/O errors
        cmd = f"echo 0 100 error | dmsetup create chaos-disk"
        subprocess.run(cmd, shell=True, capture_output=True)
        
class ApplicationChaos(ChaosMonkey):
    """Application-level chaos scenarios"""
    
    def corrupt_database_connection(self):
        """Simulate database connection issues"""
        print("Corrupting database connections")
        # Drop connections using iptables
        cmd = "iptables -A OUTPUT -p tcp --dport 5432 -j REJECT --reject-with tcp-reset"
        subprocess.run(cmd.split(), capture_output=True)
        
    def simulate_cache_failure(self):
        """Simulate cache (Redis) failure"""
        print("Simulating cache failure")
        try:
            redis_container = self.docker_client.containers.get("computehive_redis")
            redis_container.pause()
            time.sleep(30)
            redis_container.unpause()
        except docker.errors.NotFound:
            print("Redis container not found")
            
    def inject_api_errors(self, error_rate=0.1):
        """Inject random API errors"""
        print(f"Injecting {error_rate*100}% API errors")
        # This would require a proxy or service mesh to inject errors
        pass
        
class ChaosTestRunner:
    """Run chaos engineering tests"""
    
    def __init__(self):
        self.network_chaos = NetworkChaos(
            target_services=["auth", "scheduler", "marketplace"],
            chaos_level="medium"
        )
        self.service_chaos = ServiceChaos(
            target_services=["auth", "scheduler", "marketplace"],
            chaos_level="medium"
        )
        self.resource_chaos = ResourceChaos(chaos_level="light")
        self.app_chaos = ApplicationChaos(chaos_level="medium")
        
    def verify_system_health(self):
        """Check if system is still operational"""
        try:
            # Check health endpoint
            response = requests.get("http://localhost:8080/health", timeout=5)
            if response.status_code != 200:
                return False
                
            # Check critical services
            critical_endpoints = [
                "/api/v1/auth/health",
                "/api/v1/jobs/health",
                "/api/v1/marketplace/health"
            ]
            
            for endpoint in critical_endpoints:
                response = requests.get(f"http://localhost:8080{endpoint}", timeout=5)
                if response.status_code != 200:
                    return False
                    
            return True
        except:
            return False
            
    def measure_recovery_time(self, chaos_function, *args):
        """Measure how long system takes to recover from chaos"""
        start_time = time.time()
        
        # Inject chaos
        chaos_function(*args)
        
        # Wait for system to detect failure
        time.sleep(5)
        
        # Measure recovery
        recovery_start = time.time()
        while not self.verify_system_health():
            time.sleep(1)
            if time.time() - recovery_start > 300:  # 5 minute timeout
                return -1  # Recovery failed
                
        recovery_time = time.time() - recovery_start
        return recovery_time
        
    def run_chaos_scenario(self, name, chaos_function, *args, **kwargs):
        """Run a chaos scenario and collect metrics"""
        print(f"\n{'='*50}")
        print(f"Running chaos scenario: {name}")
        print(f"Time: {datetime.now()}")
        print(f"{'='*50}")
        
        # Verify system is healthy before chaos
        if not self.verify_system_health():
            print("System not healthy before chaos!")
            return False
            
        # Measure baseline performance
        baseline_response_time = self.measure_api_response_time()
        
        # Run chaos
        recovery_time = self.measure_recovery_time(chaos_function, *args)
        
        # Measure post-chaos performance
        post_chaos_response_time = self.measure_api_response_time()
        
        # Report results
        print(f"\nResults:")
        print(f"- Recovery time: {recovery_time}s")
        print(f"- Baseline response time: {baseline_response_time}ms")
        print(f"- Post-chaos response time: {post_chaos_response_time}ms")
        print(f"- Performance degradation: {(post_chaos_response_time/baseline_response_time - 1)*100:.1f}%")
        
        return recovery_time > 0 and recovery_time < 60  # Success if recovered within 1 minute
        
    def measure_api_response_time(self):
        """Measure average API response time"""
        times = []
        for _ in range(10):
            start = time.time()
            try:
                requests.get("http://localhost:8080/api/v1/jobs", timeout=5)
                times.append((time.time() - start) * 1000)
            except:
                pass
        return sum(times) / len(times) if times else 999999

# Test scenarios
class TestChaosEngineering:
    """Chaos engineering test cases"""
    
    @pytest.fixture
    def chaos_runner(self):
        return ChaosTestRunner()
        
    def test_network_packet_loss(self, chaos_runner):
        """Test system resilience to packet loss"""
        success = chaos_runner.run_chaos_scenario(
            "10% Packet Loss",
            chaos_runner.network_chaos.simulate_packet_loss,
            "eth0", 10
        )
        assert success, "System failed to handle packet loss"
        
    def test_service_failure(self, chaos_runner):
        """Test system resilience to service failures"""
        success = chaos_runner.run_chaos_scenario(
            "Random Service Kill",
            chaos_runner.service_chaos.kill_random_service
        )
        assert success, "System failed to handle service failure"
        
    def test_cpu_stress(self, chaos_runner):
        """Test system under CPU stress"""
        success = chaos_runner.run_chaos_scenario(
            "CPU Stress (2 cores, 30s)",
            chaos_runner.resource_chaos.consume_cpu,
            2, 30
        )
        assert success, "System failed under CPU stress"
        
    def test_memory_pressure(self, chaos_runner):
        """Test system under memory pressure"""
        success = chaos_runner.run_chaos_scenario(
            "Memory Pressure (2GB, 30s)",
            chaos_runner.resource_chaos.consume_memory,
            2000, 30
        )
        assert success, "System failed under memory pressure"
        
    def test_database_failure(self, chaos_runner):
        """Test system resilience to database issues"""
        success = chaos_runner.run_chaos_scenario(
            "Database Connection Failure",
            chaos_runner.app_chaos.corrupt_database_connection
        )
        assert success, "System failed to handle database issues"
        
    def test_cascading_failure(self, chaos_runner):
        """Test system resilience to cascading failures"""
        def cascading_chaos():
            # Kill auth service
            chaos_runner.service_chaos.kill_random_service()
            time.sleep(5)
            # Add network delay
            chaos_runner.network_chaos.simulate_network_delay("eth0", 200)
            time.sleep(5)
            # Consume CPU
            chaos_runner.resource_chaos.consume_cpu(1, 20)
            
        success = chaos_runner.run_chaos_scenario(
            "Cascading Failure",
            cascading_chaos
        )
        assert success, "System failed to handle cascading failures"

if __name__ == "__main__":
    runner = ChaosTestRunner()
    
    # Run a sample chaos scenario
    runner.run_chaos_scenario(
        "Sample Network Chaos",
        runner.network_chaos.simulate_packet_loss,
        "eth0", 5
    ) 