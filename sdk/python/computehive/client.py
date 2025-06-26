"""
ComputeHive Python SDK
A Python client library for interacting with the ComputeHive distributed compute platform.
"""

import os
import time
import json
import hashlib
from typing import Dict, List, Optional, Any, Callable
from dataclasses import dataclass, asdict
from enum import Enum
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry


class JobType(Enum):
    """Supported job types"""
    DOCKER = "docker"
    BINARY = "binary"
    SCRIPT = "script"
    WASM = "wasm"


class JobStatus(Enum):
    """Job status values"""
    PENDING = "pending"
    SCHEDULED = "scheduled"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class ResourceRequirements:
    """Resource requirements for a job"""
    cpu_cores: int
    memory_mb: int
    gpu_count: int = 0
    gpu_type: Optional[str] = None
    storage_mb: int = 1024
    network_mbps: int = 100
    trusted_exec: bool = False
    capabilities: Optional[List[str]] = None

    def to_dict(self) -> Dict:
        data = asdict(self)
        # Remove None values
        return {k: v for k, v in data.items() if v is not None}


@dataclass
class SLARequirements:
    """Service Level Agreement requirements"""
    max_latency_ms: Optional[int] = None
    min_availability: Optional[float] = None
    max_cost_per_hour: Optional[float] = None
    preferred_regions: Optional[List[str]] = None

    def to_dict(self) -> Dict:
        data = asdict(self)
        return {k: v for k, v in data.items() if v is not None}


@dataclass
class JobPayload:
    """Job execution payload"""
    # Docker job fields
    image: Optional[str] = None
    command: Optional[List[str]] = None
    env: Optional[List[str]] = None
    
    # Binary job fields
    binary_url: Optional[str] = None
    args: Optional[List[str]] = None
    
    # Script job fields
    script: Optional[str] = None
    language: Optional[str] = None
    
    # Input/output
    input_data: Optional[str] = None
    output_path: Optional[str] = None

    def to_dict(self) -> Dict:
        data = asdict(self)
        return {k: v for k, v in data.items() if v is not None}


class ComputeHiveError(Exception):
    """Base exception for ComputeHive SDK"""
    pass


class AuthenticationError(ComputeHiveError):
    """Authentication related errors"""
    pass


class JobError(ComputeHiveError):
    """Job related errors"""
    pass


class ComputeHiveClient:
    """Main client for interacting with ComputeHive"""
    
    def __init__(
        self,
        api_url: Optional[str] = None,
        api_key: Optional[str] = None,
        timeout: int = 30,
        max_retries: int = 3
    ):
        """
        Initialize ComputeHive client
        
        Args:
            api_url: Base URL for ComputeHive API
            api_key: API key for authentication
            timeout: Request timeout in seconds
            max_retries: Maximum number of retries for failed requests
        """
        self.api_url = api_url or os.getenv("COMPUTEHIVE_API_URL", "https://api.computehive.io")
        self.api_key = api_key or os.getenv("COMPUTEHIVE_API_KEY")
        
        if not self.api_key:
            raise AuthenticationError("API key is required. Set COMPUTEHIVE_API_KEY environment variable or pass api_key parameter.")
        
        # Setup session with retry logic
        self.session = requests.Session()
        retry_strategy = Retry(
            total=max_retries,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504],
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)
        
        # Set default headers
        self.session.headers.update({
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
            "User-Agent": "ComputeHive-Python-SDK/1.0.0"
        })
        
        self.timeout = timeout
    
    def _make_request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict] = None,
        params: Optional[Dict] = None
    ) -> Dict:
        """Make HTTP request to API"""
        url = f"{self.api_url}{endpoint}"
        
        try:
            response = self.session.request(
                method=method,
                url=url,
                json=data,
                params=params,
                timeout=self.timeout
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 401:
                raise AuthenticationError("Invalid API key or authentication failed")
            elif e.response.status_code == 400:
                raise ComputeHiveError(f"Bad request: {e.response.text}")
            else:
                raise ComputeHiveError(f"API request failed: {e}")
        except requests.exceptions.RequestException as e:
            raise ComputeHiveError(f"Network error: {e}")
    
    def submit_job(
        self,
        job_type: JobType,
        requirements: ResourceRequirements,
        payload: JobPayload,
        priority: int = 5,
        timeout: int = 3600,
        max_retries: int = 3,
        sla_requirements: Optional[SLARequirements] = None
    ) -> Dict:
        """
        Submit a new compute job
        
        Args:
            job_type: Type of job (docker, binary, script, wasm)
            requirements: Resource requirements for the job
            payload: Job execution payload
            priority: Job priority (0-10, higher is more important)
            timeout: Job timeout in seconds
            max_retries: Maximum retries if job fails
            sla_requirements: Optional SLA requirements
            
        Returns:
            Job details including job ID
        """
        data = {
            "type": job_type.value,
            "requirements": requirements.to_dict(),
            "payload": payload.to_dict(),
            "priority": priority,
            "timeout": timeout,
            "max_retries": max_retries
        }
        
        if sla_requirements:
            data["sla_requirements"] = sla_requirements.to_dict()
        
        return self._make_request("POST", "/api/v1/jobs", data=data)
    
    def get_job(self, job_id: str) -> Dict:
        """Get job details by ID"""
        return self._make_request("GET", f"/api/v1/jobs/{job_id}")
    
    def list_jobs(
        self,
        status: Optional[JobStatus] = None,
        limit: int = 100,
        offset: int = 0
    ) -> List[Dict]:
        """
        List jobs
        
        Args:
            status: Filter by job status
            limit: Maximum number of jobs to return
            offset: Offset for pagination
            
        Returns:
            List of job details
        """
        params = {"limit": limit, "offset": offset}
        if status:
            params["status"] = status.value
        
        return self._make_request("GET", "/api/v1/jobs", params=params)
    
    def cancel_job(self, job_id: str) -> None:
        """Cancel a job"""
        self._make_request("POST", f"/api/v1/jobs/{job_id}/cancel")
    
    def get_job_result(self, job_id: str) -> Dict:
        """Get job execution result"""
        return self._make_request("GET", f"/api/v1/jobs/{job_id}/result")
    
    def get_job_logs(self, job_id: str, tail: int = 100) -> str:
        """Get job execution logs"""
        response = self._make_request("GET", f"/api/v1/jobs/{job_id}/logs", params={"tail": tail})
        return response.get("logs", "")
    
    def wait_for_job(
        self,
        job_id: str,
        timeout: int = 3600,
        poll_interval: int = 5,
        callback: Optional[Callable[[Dict], None]] = None
    ) -> Dict:
        """
        Wait for job completion
        
        Args:
            job_id: Job ID to wait for
            timeout: Maximum time to wait in seconds
            poll_interval: Interval between status checks
            callback: Optional callback function called on each status update
            
        Returns:
            Final job details
        """
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            job = self.get_job(job_id)
            
            if callback:
                callback(job)
            
            status = job.get("status")
            if status in ["completed", "failed", "cancelled"]:
                return job
            
            time.sleep(poll_interval)
        
        raise JobError(f"Job {job_id} did not complete within {timeout} seconds")
    
    def submit_docker_job(
        self,
        image: str,
        command: List[str],
        cpu_cores: int = 1,
        memory_mb: int = 1024,
        gpu_count: int = 0,
        env: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Dict:
        """
        Convenience method to submit a Docker job
        
        Args:
            image: Docker image name
            command: Command to run in container
            cpu_cores: Number of CPU cores
            memory_mb: Memory in MB
            gpu_count: Number of GPUs
            env: Environment variables
            **kwargs: Additional arguments passed to submit_job
            
        Returns:
            Job details
        """
        env_list = [f"{k}={v}" for k, v in (env or {}).items()]
        
        requirements = ResourceRequirements(
            cpu_cores=cpu_cores,
            memory_mb=memory_mb,
            gpu_count=gpu_count
        )
        
        payload = JobPayload(
            image=image,
            command=command,
            env=env_list
        )
        
        return self.submit_job(
            job_type=JobType.DOCKER,
            requirements=requirements,
            payload=payload,
            **kwargs
        )
    
    def submit_script_job(
        self,
        script: str,
        language: str,
        cpu_cores: int = 1,
        memory_mb: int = 1024,
        **kwargs
    ) -> Dict:
        """
        Convenience method to submit a script job
        
        Args:
            script: Script content
            language: Script language (python, javascript, bash, etc.)
            cpu_cores: Number of CPU cores
            memory_mb: Memory in MB
            **kwargs: Additional arguments passed to submit_job
            
        Returns:
            Job details
        """
        requirements = ResourceRequirements(
            cpu_cores=cpu_cores,
            memory_mb=memory_mb
        )
        
        payload = JobPayload(
            script=script,
            language=language
        )
        
        return self.submit_job(
            job_type=JobType.SCRIPT,
            requirements=requirements,
            payload=payload,
            **kwargs
        )
    
    def get_agents(self, active_only: bool = True) -> List[Dict]:
        """Get list of available compute agents"""
        params = {"active": active_only}
        return self._make_request("GET", "/api/v1/agents", params=params)
    
    def get_agent(self, agent_id: str) -> Dict:
        """Get agent details by ID"""
        return self._make_request("GET", f"/api/v1/agents/{agent_id}")
    
    def get_marketplace_offers(
        self,
        min_cpu: Optional[int] = None,
        min_memory: Optional[int] = None,
        max_price: Optional[float] = None,
        region: Optional[str] = None
    ) -> List[Dict]:
        """
        Get marketplace compute offers
        
        Args:
            min_cpu: Minimum CPU cores
            min_memory: Minimum memory in MB
            max_price: Maximum price per hour
            region: Preferred region
            
        Returns:
            List of compute offers
        """
        params = {}
        if min_cpu:
            params["min_cpu"] = min_cpu
        if min_memory:
            params["min_memory"] = min_memory
        if max_price:
            params["max_price"] = max_price
        if region:
            params["region"] = region
        
        return self._make_request("GET", "/api/v1/marketplace/offers", params=params)
    
    def get_billing_info(self) -> Dict:
        """Get billing information for the account"""
        return self._make_request("GET", "/api/v1/billing")
    
    def get_usage_stats(
        self,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None
    ) -> Dict:
        """
        Get usage statistics
        
        Args:
            start_date: Start date (YYYY-MM-DD)
            end_date: End date (YYYY-MM-DD)
            
        Returns:
            Usage statistics
        """
        params = {}
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        
        return self._make_request("GET", "/api/v1/stats/usage", params=params)


# Convenience functions
def create_client(api_key: Optional[str] = None) -> ComputeHiveClient:
    """Create a ComputeHive client instance"""
    return ComputeHiveClient(api_key=api_key)


def quick_run_docker(
    image: str,
    command: List[str],
    api_key: Optional[str] = None,
    wait: bool = True,
    **kwargs
) -> Dict:
    """
    Quick helper to run a Docker job
    
    Args:
        image: Docker image
        command: Command to run
        api_key: API key (uses env var if not provided)
        wait: Whether to wait for completion
        **kwargs: Additional job parameters
        
    Returns:
        Job result if wait=True, job details otherwise
    """
    client = create_client(api_key)
    job = client.submit_docker_job(image, command, **kwargs)
    
    if wait:
        print(f"Job {job['id']} submitted, waiting for completion...")
        result = client.wait_for_job(job['id'])
        if result['status'] == 'completed':
            return client.get_job_result(job['id'])
        else:
            raise JobError(f"Job failed with status: {result['status']}")
    
    return job 