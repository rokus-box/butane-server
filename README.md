![image](https://github.com/rokus-box/butane-server/assets/46532725/bd6f7eed-636c-471f-8fab-40c7a0798e7b)

# Butane

This repository contains the source code for a language-agnostic password manager app implemented utilizing AWS, Terraform, Golang, and Node.js.

## What I aimed to learn and demonstrate.

- **Language Agnostic:** Demonstrated the use of Golang and Node.js (and other languages potentially such as Python, Java, C# and Rust) to build a language-agnostic application.
- **AWS Lambda:** Utilized AWS Lambda for serverless architecture, ensuring scalability and cost-effectiveness and also language agnosticism.
- **DynamoDB Single Table Schema:** Implemented a single table schema design in DynamoDB to manage various relations efficiently.
- **Infrastructure as Code (IaC):** Used Terraform for Infrastructure as Code (IaC), providing a reproducible and version-controlled infrastructure setup. Used S3 to store latest state and DynamoDB for the lock mechanism.
- **Continuous Integration/Continuous Deployment (CI/CD):** Set up a CI/CD pipeline using GitHub Actions for automated testing and deployment.
- **OIDC:** Made CI/CD pipeline using ephemeral access keys by setting up OIDC provider in AWS to avoid using hard-coded secret keys.
- **OAuth2:** Handled authorization using 3rd party OAuth2 clients to avoid spam accounts as much as possible by allowing only verified users.
