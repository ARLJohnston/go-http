import http from 'k6/http';
import { sleep } from 'k6';
import { Client } from 'k6/net/grpc';


export const options = {
  scenarios: {
    grpcUser: {
      executor: 'constant-vus',
      exec: 'grpcUser',
      vus: 10,
      duration: '30s',
    },
    httpUser: {
      executor: 'constant-vus',
      exec: 'httpUser',
      vus: 10,
      duration: '30s',
    },
  },

};

export function grpcUser() {
  const client = new Client();
  client.connect('127.0.0.1:50051', { reflect: true });

  const response = client.invoke('album.Albums/Read', null);

  check(response, {
    'status is OK': (r) => r && r.status === StatusOK,
  });
}

export function httpUser() {
  http.get('http://127.0.0.1:3000/metrics');
}
