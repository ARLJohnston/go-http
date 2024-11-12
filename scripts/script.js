import http from 'k6/http';
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';
import { Client } from 'k6/net/grpc';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const images = ['https://raw.githubusercontent.com/msikma/pokesprite/refs/heads/master/icons/pokemon/regular/abomasnow-mega.png']

export const options = {
  scenarios: {
    grpcCreator: {
      executor: 'constant-vus',
      exec: 'grpcCreate',
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


export function grpcCreate() {
	client.connect('127.0.0.1:50051', {plaintext: true});

  const data = {id: 0, title: 'Title', artist: 'Artist', price: 12.99, cover: 'https://raw.githubusercontent.com/msikma/pokesprite/refs/heads/master/icons/pokemon/regular/abomasnow-mega.png'};

  const response = client.invoke('album.Albums/Create', data);

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });

	client.close()
}

export function grpcRead() {
  client.connect('127.0.0.1:50051', {plaintext: true});

  const data = {};
	const response = client.invoke('album.Albums/Read', data);

  check(response, {
    'status is OK': (r) => r && r.status === grpc.StatusOK,
  });
	client.close()
}

export function httpUser() {
  http.get('http://127.0.0.1:3000/metrics');
}

export function httpMetricer() {
  http.get('http://127.0.0.1:3000/');
}
