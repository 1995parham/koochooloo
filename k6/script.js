import http from 'k6/http';
import { check, group } from 'k6';

const baseURL = "http://127.0.0.1:1378"

export default function() {
  group("healthz", () => {
    let res = http.get(`${baseURL}/healthz`);

    check(res, {
      "success": (res) => res.status === 204,
    });
  });
  group("short", () => {
    let name = ""
    group("create", () => {
      let payload = JSON.stringify({
        "url": "https://elahe-dastan.github.io",
      });

      let res = http.post(`${baseURL}/api/urls`, payload, {
        headers: {
          "Content-Type": "application/json",
        }
      })

      check(res, {
        "success": (res) => res.status == 200,
      })

      name = res.json()
    })

    console.log(name)

    group("fetch", () => {
      let res = http.get(`${baseURL}/api/${name}`)

      check(res, {
        "success": (res) => res.status == 200,
      })
    })

  })
}
