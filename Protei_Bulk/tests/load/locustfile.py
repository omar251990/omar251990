"""
Locust Load Testing Configuration
Tests Protei_Bulk platform at scale (10,000+ TPS target)
"""

from locust import HttpUser, task, between, events
from locust.contrib.fasthttp import FastHttpUser
import random
import string
import json
from datetime import datetime


class MessagingUser(FastHttpUser):
    """
    Simulates a user sending messages through the API
    Uses FastHttpUser for better performance
    """

    # Wait time between tasks (simulates user behavior)
    wait_time = between(0.1, 0.5)  # Very short for load testing

    # API credentials
    api_key = "ADMIN_API_KEY_test"  # Replace with actual API key

    def on_start(self):
        """Called when a user starts - authenticate"""
        # Set API key header
        self.client.headers.update({"X-API-Key": self.api_key})

        # Test authentication
        response = self.client.get("/api/v1/health")
        if response.status_code != 200:
            print(f"Health check failed: {response.status_code}")

    @task(10)  # Weight: 10 (most common operation)
    def send_single_message(self):
        """Send a single SMS message"""
        payload = {
            "from": "LoadTest",
            "to": self.generate_phone_number(),
            "text": f"Load test message at {datetime.now().isoformat()}",
            "encoding": "GSM7",
            "priority": random.choice(["NORMAL", "HIGH"])
        }

        with self.client.post(
            "/api/v1/messages",
            json=payload,
            catch_response=True,
            name="Send Message"
        ) as response:
            if response.status_code == 201:
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")

    @task(5)  # Weight: 5
    def send_bulk_messages(self):
        """Send bulk messages"""
        count = random.randint(10, 100)
        messages = [
            {
                "to": self.generate_phone_number(),
                "text": f"Bulk message {i}"
            }
            for i in range(count)
        ]

        payload = {
            "from": "LoadTest",
            "messages": messages,
            "priority": "NORMAL"
        }

        with self.client.post(
            "/api/v1/messages/bulk",
            json=payload,
            catch_response=True,
            name="Send Bulk Messages"
        ) as response:
            if response.status_code in [200, 201]:
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")

    @task(3)  # Weight: 3
    def get_message_status(self):
        """Query message status"""
        # Generate a random message ID (in real scenario, track actual IDs)
        message_id = f"msg_{self.generate_random_string(16)}"

        with self.client.get(
            f"/api/v1/messages/{message_id}",
            catch_response=True,
            name="Get Message Status"
        ) as response:
            if response.status_code in [200, 404]:  # 404 is expected for random IDs
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")

    @task(2)  # Weight: 2
    def list_messages(self):
        """List recent messages"""
        params = {
            "page": 1,
            "limit": 50,
            "status": random.choice(["PENDING", "SENT", "DELIVERED", "FAILED"])
        }

        with self.client.get(
            "/api/v1/messages",
            params=params,
            catch_response=True,
            name="List Messages"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")

    @task(1)  # Weight: 1
    def create_campaign(self):
        """Create a campaign"""
        payload = {
            "name": f"Load Test Campaign {self.generate_random_string(8)}",
            "sender_id": "LoadTest",
            "message_content": "Campaign message from load test",
            "recipient_type": "MANUAL",
            "total_recipients": random.randint(100, 1000),
            "schedule_type": "IMMEDIATE",
            "priority": "NORMAL"
        }

        with self.client.post(
            "/api/v1/campaigns",
            json=payload,
            catch_response=True,
            name="Create Campaign"
        ) as response:
            if response.status_code == 201:
                response.success()
            else:
                response.failure(f"Failed with status {response.status_code}")

    @staticmethod
    def generate_phone_number():
        """Generate a random phone number"""
        return "98765" + ''.join(random.choices(string.digits, k=5))

    @staticmethod
    def generate_random_string(length):
        """Generate random alphanumeric string"""
        return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))


class SMPPUser(FastHttpUser):
    """
    Simulates SMPP connections and message submission
    For SMPP protocol testing
    """

    wait_time = between(0.05, 0.2)  # Very fast for SMPP

    @task
    def smpp_submit(self):
        """Simulate SMPP submit_sm"""
        # In reality, this would use SMPP protocol
        # Here we simulate via HTTP for load testing purposes
        payload = {
            "from": "SMPP",
            "to": MessagingUser.generate_phone_number(),
            "text": "SMPP load test message",
            "source_type": "SMPP"
        }

        with self.client.post(
            "/api/v1/messages",
            json=payload,
            catch_response=True,
            name="SMPP Submit"
        ) as response:
            if response.status_code in [200, 201]:
                response.success()


# Event listeners for custom metrics
@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    """Called when load test starts"""
    print(f"\n{'='*60}")
    print(f"Protei_Bulk Load Test Starting")
    print(f"Target: {environment.host}")
    print(f"{'='*60}\n")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Called when load test stops"""
    stats = environment.stats

    print(f"\n{'='*60}")
    print(f"Protei_Bulk Load Test Complete")
    print(f"{'='*60}")
    print(f"Total Requests: {stats.total.num_requests}")
    print(f"Total Failures: {stats.total.num_failures}")
    print(f"Average Response Time: {stats.total.avg_response_time:.2f}ms")
    print(f"Median Response Time: {stats.total.median_response_time:.2f}ms")
    print(f"95th Percentile: {stats.total.get_response_time_percentile(0.95):.2f}ms")
    print(f"99th Percentile: {stats.total.get_response_time_percentile(0.99):.2f}ms")
    print(f"Requests/sec: {stats.total.total_rps:.2f}")
    print(f"{'='*60}\n")


# Custom load shape for gradual ramp-up
from locust import LoadTestShape

class GradualRampUp(LoadTestShape):
    """
    Gradually ramp up load to test system scalability
    Simulates realistic traffic growth
    """

    stages = [
        # (duration, users, spawn_rate)
        (60, 100, 10),      # Ramp to 100 users in 60 seconds
        (120, 500, 20),     # Ramp to 500 users in next 60 seconds
        (180, 1000, 50),    # Ramp to 1000 users
        (240, 2000, 100),   # Ramp to 2000 users
        (300, 5000, 200),   # Ramp to 5000 users (target for 10K TPS)
        (360, 5000, 0),     # Hold at 5000 users for 60 seconds
        (420, 0, 500),      # Ramp down
    ]

    def tick(self):
        run_time = self.get_run_time()

        for stage_duration, users, spawn_rate in self.stages:
            if run_time < stage_duration:
                return (users, spawn_rate)

        return None  # End test


if __name__ == "__main__":
    import os
    os.system("locust -f locustfile.py --host=http://localhost:8080")
