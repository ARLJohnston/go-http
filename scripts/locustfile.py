import time
import random
import queue
from locust import FastHttpUser, task, between, events
import grpc
import grpc_user
import album_pb2
import album_pb2_grpc
from faker import Faker

Faker.seed(4321)
fake = Faker()
random.seed(8765)


@events.init_command_line_parser.add_listener
def _(parser):
    parser.add_argument(
        "--host-addr", type=str, default="localhost", help="IP address of the cluster"
    )

class APIUser(grpc_user.GrpcUser):
    """Simulation of user using a developer API to directly call gRPC methods"""
    stub_class = album_pb2_grpc.AlbumsStub
    offset = 6

    @property
    def host(self):
        return self.environment.parsed_options.host_addr + ":50051"

    @task
    def createDelete(self):
        album = album_pb2.Album(
            id=self.offset,
            title=fake.bs(),
            artist=fake.name(),
            score=random.randint(-50, 50),
            cover=fake.image_url(),
        )
        self.stub.Create(album)
        self.offset += 1

        self.stub.Delete(album_pb2.Identifier(id=album))


class NormalUser(FastHttpUser):
    """Simulation of user using http to access page and vote on posts"""
    network_timeout = 3.0
    connection_timeout = 3.0
    header = {
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
    }

    @property
    def host(self):
        return "http://" + self.environment.parsed_options.host_addr + ":8000"

    @task
    def get_front_end(self):
        self.client.get("/", header=self.header)

    @task
    def get_metrics(self):
        self.client.get("/metrics", header=self.header)

    @task
    def vote(self):
        id = random.randint(1, 5)  # Choose one of the pre-existing albums
        self.client.post("/post", {random.choice(["up", "down"]): id})
