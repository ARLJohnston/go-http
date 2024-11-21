import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { vu, scenario } from 'k6/execution';
import { check, sleep } from 'k6';
import { Client, Stream } from 'k6/net/grpc';
import { Counter } from 'k6/metrics';
import faker from "k6/x/faker";


export const options = {
  scenarios: {
    grpcCreatorDeletor: {
      executor: 'constant-vus',
      exec: 'grpcCreateDelete',
      vus: 10,
      duration: '30s',
    },
    grpcReader: {
      executor: 'constant-vus',
      exec: 'grpcRead',
      vus: 10,
      duration: '30s',
    },
    httpUser: {
      executor: 'constant-vus',
      exec: 'httpUser',
      vus: 10,
      duration: '30s',
    },
    httpMetrics: {
      executor: 'constant-vus',
      exec: 'httpMetricer',
      vus: 10,
      duration: '30s',
    },
  },

};

const client = new Client();
client.load(['../proto/'], 'album.proto');


export function grpcCreateDelete() {
  if (__ITER == 0) {
    client.connect('127.0.0.1:50051', { plaintext: true });
  }

  const imageURL = faker.internet.imageUrl(500,500);
  const movie = faker.movie.movie();

  const id = vu.idInTest + 10;

  const data = {
      id: id,
      title: movie.name,
      artist: movie.genre,
      price: faker.payment.price(0,100),
      cover: imageURL
  };

  const response = client.invoke('album.Albums/Create', data);

  // Update here

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });

  sleep(0.1);

  const deleteResponse = client.invoke('album.Albums/Delete', data);

  check(deleteResponse, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });
}

export function grpcRead() {
  if (__ITER == 0) {
    client.connect('127.0.0.1:50051', { plaintext: true });
  }
  const stream = new Stream(client, 'album.Albums/Read')

  stream.on('data', function () {
    console.log('Data');
  });

  stream.on("error", (e) => {
    console.log("Error: " + JSON.stringify(e));
    stream.end();
  });
}

export function httpUser() {
  const response = http.get('http://127.0.0.1:3000/');
  check(response, {
    'is status 200': (r) => r.status === 200,
  });
}

export function httpMetricer() {
  const response = http.get('http://127.0.0.1:3000/metrics');
  check(response, {
    'is status 200': (r) => r.status === 200,
  });
}
