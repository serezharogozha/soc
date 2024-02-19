import http from 'k6/http';
import { check } from 'k6';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

export const options = {
    vus: 1,
    duration: '30s'
};

export default function () {
    let min = 100;
    let max = 50000;
    let randomUser = Math.floor(Math.random() * (max - min + 1)) + min;
    const url = 'http://localhost:8080/dialog/'+ randomUser +'/list';
    const params = {
        headers: {
            'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgzNzI0OTksInVzZXJfaWQiOjF9.LZ5kEYMSqecWGylmul-xgOF_dJBomX-tdDkhbf2if_U',
        },
    };


    let res = http.get(url, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response body is not empty': (r) => r.body.length > 0,
    });
}