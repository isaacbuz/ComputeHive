import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export let options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 500 },   // Ramp up to 500 users
    { duration: '5m', target: 500 },   // Stay at 500 users
    { duration: '2m', target: 1000 },  // Ramp up to 1000 users
    { duration: '5m', target: 1000 },  // Stay at 1000 users
    { duration: '2m', target: 0 },     // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],  // 99% of requests under 500ms
    http_req_failed: ['rate<0.01'],    // Error rate under 1%
    errors: ['rate<0.01'],             // Custom error rate under 1%
  },
};

const BASE_URL = __ENV.BASE_URL || 'https://api.computehive.io';

// Helper function to generate test data
function generateJobData() {
  return {
    type: 'compute',
    name: `test-job-${Date.now()}`,
    resources: {
      cpu: Math.floor(Math.random() * 8) + 1,
      memory: `${Math.floor(Math.random() * 16) + 1}GB`,
      gpu: Math.random() > 0.7 ? 1 : 0,
    },
    docker_image: 'ubuntu:latest',
    command: 'echo "Hello from ComputeHive"',
    timeout: 3600,
  };
}

// Test scenarios
export default function() {
  // Scenario 1: Authentication
  let authResponse = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: 'test@computehive.io',
    password: 'testPassword123!',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(authResponse, {
    'login successful': (r) => r.status === 200,
    'received auth token': (r) => r.json('access_token') !== '',
  });

  errorRate.add(authResponse.status !== 200);

  if (authResponse.status !== 200) {
    return;
  }

  const authToken = authResponse.json('access_token');
  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Scenario 2: Submit Job
  let jobData = generateJobData();
  let jobResponse = http.post(`${BASE_URL}/jobs`, JSON.stringify(jobData), {
    headers: authHeaders,
  });

  check(jobResponse, {
    'job created': (r) => r.status === 201,
    'job id returned': (r) => r.json('id') !== '',
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(jobResponse.status !== 201);

  // Scenario 3: Get Job Status
  if (jobResponse.status === 201) {
    const jobId = jobResponse.json('id');
    sleep(1);

    let statusResponse = http.get(`${BASE_URL}/jobs/${jobId}`, {
      headers: authHeaders,
    });

    check(statusResponse, {
      'job status retrieved': (r) => r.status === 200,
      'status field exists': (r) => r.json('status') !== '',
    });

    errorRate.add(statusResponse.status !== 200);
  }

  // Scenario 4: List Jobs
  let listResponse = http.get(`${BASE_URL}/jobs?limit=10`, {
    headers: authHeaders,
  });

  check(listResponse, {
    'jobs listed': (r) => r.status === 200,
    'response is array': (r) => Array.isArray(r.json('jobs')),
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  errorRate.add(listResponse.status !== 200);

  // Scenario 5: Get Marketplace Resources
  let marketResponse = http.get(`${BASE_URL}/marketplace/resources`, {
    headers: authHeaders,
  });

  check(marketResponse, {
    'marketplace accessible': (r) => r.status === 200,
    'resources listed': (r) => r.json('resources') !== null,
  });

  errorRate.add(marketResponse.status !== 200);

  // Think time between iterations
  sleep(Math.random() * 3 + 1);
}

// Lifecycle hooks
export function setup() {
  console.log('Setting up load test...');
  
  // Verify target system is accessible
  let res = http.get(BASE_URL + '/health');
  if (res.status !== 200) {
    throw new Error(`Target system is not healthy: ${res.status}`);
  }
  
  return { startTime: new Date() };
}

export function teardown(data) {
  console.log(`Test completed. Duration: ${new Date() - data.startTime}ms`);
} 