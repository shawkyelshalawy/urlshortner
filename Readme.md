# URL Shortener Service

A lightweight URL shortener built in Go using Gin, Redis, MongoDB, and Nginx. This project demonstrates a microservices architecture with containerization (Docker) and orchestration (Kubernetes).

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Setup Instructions](#setup-instructions)
  - [Local Development with Docker Compose](#local-development-with-docker-compose)
  - [Running Unit Tests](#running-unit-tests)
- [API Usage](#api-usage)
- [Kubernetes Deployment](#kubernetes-deployment)


## Overview

The URL Shortener Service allows users to submit a long URL and receive a shortened version. When a short URL is accessed, the service redirects to the original URL. It leverages Redis for fast lookup and MongoDB for analytics logging, with Nginx acting as a reverse proxy and load balancer.

## Features

- **URL Shortening:** Submit a URL to get a shortened version.
- **Redirection:** Accessing a short URL redirects to the original.
- **Analytics Logging:** Each redirection logs access data (IP, user agent, timestamp) in MongoDB.
- **Rate Limiting:** Prevent abuse with per-IP request limits.
- **Containerized:** Built with Docker for easy deployment.
- **Orchestrated:** Deployed using Kubernetes with high availability.

## Tech Stack

- **Language:** Go (Golang)
- **Web Framework:** Gin
- **Storage:** Redis (for URL mappings) and MongoDB (for analytics)
- **Reverse Proxy:** Nginx
- **Containerization:** Docker
- **Orchestration:** Kubernetes
- **Testing:** Testify

## Setup Instructions

### Local Development with Docker Compose

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/shawkyelshalawy/urlshortner.git
   cd urlshortner
   ```

### Environment Setup

1. Create a `.env` file in the project root:

```env
# Application settings
PORT=8080
ENV=development
BASE_URL=http://nginx

# Redis connection
REDIS_ADDR=redis:6379

# MongoDB connection
MONGO_URI=mongodb://mongo:27017
MONGO_DATABASE=urlshortener
```

### Build and Start Containers:

Use Docker Compose to build images and start all containers:

```bash
docker-compose up --build
```

This will build your API image and start containers for Redis, MongoDB, Nginx, and your API.

### Test the Setup:

Visit http://localhost in your browser to see if Nginx is reverse-proxying to your API.
You can also test the health endpoint:

```bash
curl -i http://localhost/ping
```

### Running Unit Tests

Make sure you have Go installed. Then run:

```bash
go test ./...
```

This command will run all tests for the handler functions.

## API Usage

### Endpoints

#### GET /ping

Returns a JSON status response to verify the API is up.

```bash
curl -i http://localhost/ping
```

#### POST /shorten

Accepts a long URL and returns a shortened URL.
Request Body (JSON):

```json
{
  "url": "https://example.com"
}
```

Response (201 Created for new URL):

```json
{
  "short_url": "http://nginx/shortid",
  "expires_at": "2025-03-01T12:00:00Z"
}
```

If the URL already exists, it returns a 200 OK with the existing short URL.

#### GET /:shortID

Redirects to the original URL corresponding to the short ID.

```bash
curl -I http://localhost/shortid
```

The response includes a Location header with the original URL.

## Kubernetes Deployment

This project includes Kubernetes YAML files to deploy the following components:

- Namespace: All resources are grouped in the urlshortner namespace.
- Application: Deployed with 3 replicas, using readiness and liveness probes.
- Redis: Single replica with a persistent volume.
- MongoDB: Single replica with persistent storage.
- Nginx: Acts as a reverse proxy and load balancer via a ConfigMap.

### Deployment Steps

#### Start a Local Kubernetes Cluster:

Use Minikube or Kind.

```bash
minikube start
```

#### Apply Kubernetes Manifests:

Navigate to your deployment folder (e.g., deployments/k8s/) and run:

```bash
kubectl apply -f deployments/k8s/mongo-deployment.yaml
kubectl apply -f  deployments/k8s/redis-deployment.yaml
kubectl apply -f  deployments/k8s/app-deployment.yaml
kubectl apply -f  deployments/k8s/nginx-deployment.yaml
```

#### Verify Resources:

```bash
kubectl get pods -n urlshortner
kubectl get services -n urlshortner
```

#### Access the Application via Nginx:

using Minikube:

```bash
minikube service nginx-service -n urlshortner
```

This command opens your browser to the external IP/port for Nginx.

