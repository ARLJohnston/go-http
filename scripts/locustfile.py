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
        "--host-addr", type=str, default="localhost", help="Ip address of the cluster"
    )


created_albums = queue.Queue()


class APIUser(grpc_user.GrpcUser):
    """Simulation of user using a developer API to directly call gRPC methods"""

    stub_class = album_pb2_grpc.AlbumsStub
    offset = 6

    @property
    def host(self):
        return environment.parsed_options.host_addr + ":50051"

    @task
    def create(self):
        album = album_pb2.Album(
            id=self.offset,
            title=fake.bs(),
            artist=fake.name(),
            score=random.randint(-50, 50),
            cover=fake.image_url(),
        )
        self.stub.Create(album)
        created_albums.put(self.offset)
        self.offset += 1

    @task
    def delete(self):
        try:
            album = created_albums.get(block=False)
            self.stub.Delete(album_pb2.Identifier(id=album))
        except queue.Empty:
            pass


class NormalUser(FastHttpUser):
    """Simulation of user using http to access page and vote on posts"""

    wait_time = between(1, 5)
    network_timeout = 3.0
    connection_timeout = 3.0

    @property
    def host(self):
        return "http://" + self.environment.parsed_options.host_addr + ":8000"

    @task
    def get_front_end(self):
        self.client.get("/")

    @task
    def get_metrics(self):
        self.client.get("/metrics")

    @task
    def vote(self):
        id = random.randint(1, 5)  # Choose one of the pre-existing albums
        self.client.post("/post", {random.choice(["up", "down"]): id})


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """Cleanup all created albums after the test finishes."""

    print("Cleaning up all created albums...")
    with grpc.insecure_channel(host_addr + ":50051") as channel:
        stub = album_pb2_grpc.AlbumsStub(channel)
        while not created_albums.empty():
            album_id = created_albums.get(block=False)
            try:
                print(f"deleting {album_id}")
                stub.Delete(album_pb2.Identifier(id=album_id))
                print(f"Deleted album {album_id}")
            except:
                print(f"Failed to delete album {album_id}")
        print("Cleanup finished")
