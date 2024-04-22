import http from 'k6/http';
import { check } from 'k6';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

export const options = {
    vus: 10,
    duration: '120s'
};

export default function () {
    const url = 'http://localhost:8080/user/search';

    let data = JSON.stringify({
        "first_name": "SSSSS",
        "last_name": "Фамилия"
    });

    let res = http.post(url, data);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response body is not empty': (r) => r.body.length > 0,
    });
}