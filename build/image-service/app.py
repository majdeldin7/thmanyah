from fastapi import FastAPI, Request
from prometheus_client import Counter, Histogram, generate_latest, CONTENT_TYPE_LATEST
import time

app = FastAPI()

# ======== Prometheus Metrics ======== #
REQUEST_COUNT = Counter(
    "http_requests_total",
    "Total number of HTTP requests",
    ["method", "endpoint", "status"],
)

REQUEST_LATENCY = Histogram(
    "http_request_duration_seconds",
    "Histogram of HTTP request duration",
    ["method", "endpoint", "status"],
)

# Middleware to track metrics
@app.middleware("http")
async def metrics_middleware(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    duration = time.time() - start_time

    REQUEST_COUNT.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()

    REQUEST_LATENCY.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).observe(duration)

    return response


# ======== App Routes ======== #
@app.get("/info")
def info():
    return {"image": "ok", "service": "image-service working"}

# Health checks (optional)
@app.get("/live")
def live():
    return {"status": "live"}

@app.get("/ready")
def ready():
    return {"status": "ready"}


# ======== /metrics Endpoint ======== #
@app.get("/metrics")
def metrics():
    from fastapi.responses import Response
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)