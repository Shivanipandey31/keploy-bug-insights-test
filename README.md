# keploy-bug-insights-test

A small notes API (Go + MongoDB) used to test Keploy's Bug Insights end to end
on a self-hosted cluster: record its traffic on `main`, replay the recorded
tests on each pull request, and check what the Bug Insights page counts.

## Endpoints

| Method | Path | Answer |
|---|---|---|
| GET | `/health` | `{"status":"ok"}` |
| GET | `/notes` | all notes |
| POST | `/notes` | creates a note (`title` required) |
| GET | `/notes/{id}` | one note, or 404 |
| DELETE | `/notes/{id}` | 204, or 404 |

## Run it in a cluster

```sh
docker build -t notes-app:main .
kind load docker-image notes-app:main --name <cluster>
kubectl apply -f k8s.yaml
kubectl port-forward svc/notes-app -n bugtest 8080:8080
./send-requests.sh
```
