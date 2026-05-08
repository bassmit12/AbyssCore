import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const graphqlLatency = new Trend('graphql_latency');

// Target: abysscore-api.bassmit.dev (the GraphQL gateway)
const BASE_URL = __ENV.TARGET_URL || 'https://abysscore-api.bassmit.dev';

export const options = {
  stages: [
    { duration: '30s', target: 10 },   // ramp up to 10 users
    { duration: '1m',  target: 10 },   // hold at 10 users
    { duration: '30s', target: 25 },   // spike to 25
    { duration: '30s', target: 0  },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'],  // 95% of requests under 2s
    errors: ['rate<0.05'],              // error rate under 5%
  },
};

// Introspection query — lightweight, always valid
const INTROSPECTION = JSON.stringify({
  query: `{ __schema { queryType { name } } }`,
});

export default function () {
  const res = http.post(
    `${BASE_URL}/graphql`,
    INTROSPECTION,
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'graphql_introspection' },
    }
  );

  const ok = check(res, {
    'status 200': (r) => r.status === 200,
    'no errors field': (r) => {
      try {
        return !JSON.parse(r.body).errors;
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!ok);
  graphqlLatency.add(res.timings.duration);

  sleep(1);
}
