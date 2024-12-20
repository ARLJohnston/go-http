import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { vu, scenario } from 'k6/execution';
import { check, sleep } from 'k6';
import { Client, Stream } from 'k6/net/grpc';
import { Counter } from 'k6/metrics';
import faker from "k6/x/faker";

const defaultOptions = {
		vus: 10,
		duration: '30s'
};

export const options = {
  scenarios: {
    grpcCreatorDeletor: {
      executor: 'constant-vus',
      exec: 'grpcCreateDelete',
      vus: defaultOptions.vus,
      duration: defaultOptions.duration,
    },
    grpcReader: {
      executor: 'constant-vus',
      exec: 'grpcRead',
      vus: defaultOptions.vus,
      duration: defaultOptions.duration,
    },
    httpUser: {
      executor: 'constant-vus',
      exec: 'httpUser',
      vus: defaultOptions.vus,
      duration: defaultOptions.duration,
    },
    httpMetrics: {
      executor: 'constant-vus',
      exec: 'httpMetricer',
      vus: defaultOptions.vus,
      duration: defaultOptions.duration,
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

  const data = {
      id: vu.idInTest,
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

  const deleteReq = {
      id: response.id,
      title: movie.name,
      artist: movie.genre,
      price: faker.payment.price(0,100),
      cover: imageURL
  };
  sleep(0.1);

  const deleteResponse = client.invoke('album.Albums/Delete', deleteReq);

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
