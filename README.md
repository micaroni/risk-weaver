# Risk Weaver
An AI-assisted Kubernetes security analytics platform that correlates policy violations, vulnerabilities, identity exposure, and runtime anomalies into explainable workload risk assessments.

## Project Contract
### Version 0.1
The system accepts a workload and security findings, calculates a deterministic risk score, stores results, and returns an assessment with an explanation.
 #### What is "done"?
A user can submit a JSON object to the `/workloads` endpoint and receive a workload ID in response.

