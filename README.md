# thmanyah
## 🏷️ Introduction
As a **Site Reliability Engineering (SRE)** engineer, the goal of this project was to build a reliable, secure, and scalable Kubernetes environment for multiple distributed backend services.
Due to limited time, some parts were simplified. Since coding is not my primary specialty, I used **ChatGPT** to help with application code while I focused on infrastructure, deployment, reliability, and security.


## 🚀 Project Overview
This project deploys multiple backend services inside Kubernetes:
- API Service
- Authentication Service
- Image Service

Following SRE principles:
- High Availability
- Fault Tolerance
- Auto Healing
- Observability
- Zero-Trust Networking
- Resource & Capacity Management

## 🌟 Key Features
- Deployments, Services, Ingress
- HPA Autoscaling
- Network Policies (Zero Trust)
- Liveness & Readiness Probes
- Pod Disruption Budget (PDB)
- Anti-Affinity rules
- Secrets Management
- Non-root Containers
- Resource Requests & Limits

## 🧩 Architecture
Client → Ingress → API → Auth → DB → Image → S3

## 📦 Project Structure for **Api-service** for example:
api-service/
  deploy.yaml
  svc.yaml
  ingress.yaml
  hpa.yaml
  pdb.yaml
  np.yaml
  secrets.yaml

## 🛠️ Deployment Steps

```bash
kubectl create namespace api-service
kubectl apply -f regcred.yaml
kubectl apply -f secrets.yaml
kubectl apply -f deploy.yaml
kubectl apply -f svc.yaml
kubectl apply -f ingress.yaml
kubectl apply -f hpa.yaml
kubectl apply -f np.yaml
kubectl apply -f pdb.yaml
```

## 🔐 Security Hardening
Non-root execution
Drop all capabilities
Disable privilege escalation

## 🌐 Network Policies
```
Default deny
Allow ingress from: ingress controller, monitoring, auth, image
Allow egress to: DB subnet, auth, image, kube-dns
```
## 📈 Autoscaling (HPA)
CPU Target: 60%
Memory Target: 70%

## 🧪 Health Probes
Readiness + Liveness probes

## 🧱 Pod Disruption Budget
minAvailable: 1

## 🛡️ Anti-Affinity
Spread pods across nodes

## 🔐 Secrets
DB creds, AWS keys via Kubernetes Secrets

## 🌐 Ingress Access
http://api.domain.com/

## 🧠 Personal Notes
- Some parts simplified due to time constraints.
- ChatGPT helped generate some code.

## 🏁 Conclusion
This project is a production-grade Kubernetes deployment following SRE principles.
