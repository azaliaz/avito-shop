import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export let options = {
    maxVus: 10,
    target: 100, 
    duration: '30s',
    thresholds: {
        http_req_duration: ['p(99.99)<50'],
        http_req_failed: ['rate<0.0001'],
    },
};

const tokens = [
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.Cu0jgoMXBW7FiljBNqT63i0TKnwWUFBfTcTDJiMXWOg',
];

const users = [
    'user_2',
];

export default function () {
    let params = {
        headers: {
            'Authorization': `Bearer ${randomItem(tokens)}`,
            'Content-Type': 'application/json',
        },
    };

    let payload = JSON.stringify({
        toUser: randomItem(users),
        amount: '1',
    });

    let res = http.post('http://localhost:8080/api/sendCoin', payload, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
    });
}