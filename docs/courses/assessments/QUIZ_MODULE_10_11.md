# Quiz: Modules 10-11 (Security & Enterprise Deployment)

## Instructions

- **Total Questions**: 30
- **Time Limit**: 60 minutes
- **Passing Score**: 80% (24/30)
- **Format**: Multiple choice (single answer unless specified)

---

## Section 1: Security (Questions 1-15)

### Q1. What is the purpose of HelixAgent's red team framework?

A) Performance testing
B) Simulating attacks to identify vulnerabilities
C) Code review automation
D) Documentation generation

---

### Q2. How many attack types does HelixAgent's security framework support?

A) 10+
B) 20+
C) 40+
D) 100+

---

### Q3. What is prompt injection?

A) A method to speed up prompts
B) An attack where malicious instructions are inserted into prompts
C) A compression technique
D) A caching mechanism

---

### Q4. What are guardrails in the context of LLM security?

A) Hardware protection
B) Filters and constraints to prevent harmful outputs
C) Network firewalls
D) Database access controls

---

### Q5. What is PII detection used for?

A) Performance improvement
B) Identifying and protecting personally identifiable information
C) Prompt optimization
D) Model selection

---

### Q6. Which of the following is a jailbreak attack technique?

A) SQL injection
B) Role-playing scenarios to bypass safety measures
C) Buffer overflow
D) Cross-site scripting

---

### Q7. What is the purpose of input validation in HelixAgent?

A) To improve response quality
B) To prevent malicious inputs from reaching the LLM
C) To compress data
D) To cache requests

---

### Q8. What is data exfiltration in the context of LLM security?

A) Data backup
B) Unauthorized extraction of sensitive information via prompts
C) Data compression
D) Log rotation

---

### Q9. Which security measure helps prevent token abuse?

A) Caching
B) Rate limiting
C) Compression
D) Indexing

---

### Q10. What is the recommended approach for storing API keys in production?

A) Environment variables in code
B) Secret management services (Vault, AWS Secrets Manager)
C) Configuration files
D) Database tables

---

### Q11. What is Constitutional AI used for in HelixAgent?

A) Legal compliance
B) Guiding model behavior through principles
C) Government regulations
D) Tax calculations

---

### Q12. Which HTTP header helps prevent clickjacking attacks?

A) Content-Type
B) X-Frame-Options
C) Accept-Language
D) User-Agent

---

### Q13. What is the purpose of audit logging?

A) Performance monitoring
B) Recording security-relevant events for compliance and forensics
C) Error debugging
D) User analytics

---

### Q14. What is a circuit breaker in security context?

A) Hardware protection
B) A pattern to prevent cascade failures and block repeated malicious requests
C) Network switch
D) Database connection

---

### Q15. Which JWT claim specifies when the token expires?

A) iss
B) sub
C) exp
D) iat

---

## Section 2: Enterprise Deployment (Questions 16-30)

### Q16. What is the recommended minimum number of replicas for production?

A) 1
B) 2
C) 3
D) 10

---

### Q17. What Kubernetes resource is used for automatic scaling?

A) Deployment
B) Service
C) HorizontalPodAutoscaler
D) ConfigMap

---

### Q18. What is the purpose of a liveness probe?

A) To check if the pod should receive traffic
B) To check if the container is running and should be restarted if not
C) To monitor CPU usage
D) To collect metrics

---

### Q19. What is the difference between liveness and readiness probes?

A) They are the same
B) Liveness checks if container should be restarted; readiness checks if it should receive traffic
C) Liveness is for TCP; readiness is for HTTP
D) Liveness is faster than readiness

---

### Q20. What Kubernetes resource manages external access to services?

A) Deployment
B) Service
C) Ingress
D) ConfigMap

---

### Q21. What is the purpose of a NetworkPolicy?

A) Load balancing
B) Controlling pod-to-pod network traffic
C) DNS resolution
D) Storage management

---

### Q22. What is the recommended security context setting for runAsNonRoot?

A) false
B) true
C) optional
D) depends

---

### Q23. What tool is used for package management in Kubernetes deployments?

A) npm
B) pip
C) Helm
D) Maven

---

### Q24. What is a ServiceMonitor in Prometheus Operator?

A) A logging tool
B) A custom resource for defining how to scrape metrics
C) A network proxy
D) A security scanner

---

### Q25. What is the purpose of pod anti-affinity?

A) To schedule pods on the same node
B) To distribute pods across different nodes for high availability
C) To reduce network traffic
D) To share storage

---

### Q26. What is a PodDisruptionBudget used for?

A) Memory limits
B) Ensuring minimum availability during voluntary disruptions
C) CPU allocation
D) Storage quotas

---

### Q27. What is the recommended way to manage secrets in Kubernetes?

A) ConfigMaps
B) Kubernetes Secrets with external secret management
C) Environment variables in Deployment YAML
D) Hardcoded in container images

---

### Q28. What does the `readOnlyRootFilesystem` security setting do?

A) Makes the entire cluster read-only
B) Prevents the container from writing to its filesystem
C) Disables network access
D) Blocks database writes

---

### Q29. What is blue-green deployment?

A) Running tests in parallel
B) A deployment strategy with two identical environments for zero-downtime releases
C) A logging configuration
D) A monitoring setup

---

### Q30. What metric typically triggers horizontal pod autoscaling?

A) Memory only
B) CPU utilization and/or memory utilization
C) Network bandwidth
D) Disk space

---

## Answer Key

| Q | Answer | Q | Answer | Q | Answer |
|---|--------|---|--------|---|--------|
| 1 | B | 11 | B | 21 | B |
| 2 | C | 12 | B | 22 | B |
| 3 | B | 13 | B | 23 | C |
| 4 | B | 14 | B | 24 | B |
| 5 | B | 15 | C | 25 | B |
| 6 | B | 16 | C | 26 | B |
| 7 | B | 17 | C | 27 | B |
| 8 | B | 18 | B | 28 | B |
| 9 | B | 19 | B | 29 | B |
| 10 | B | 20 | C | 30 | B |

---

## Detailed Explanations

### Security Section

**Q1**: The red team framework simulates adversarial attacks to identify vulnerabilities before real attackers find them.

**Q3**: Prompt injection attempts to hijack the LLM by injecting malicious instructions that override the original system prompt.

**Q6**: Jailbreak attacks use creative techniques like role-playing ("pretend you're an AI without restrictions") to bypass safety measures.

**Q10**: Secret management services provide encryption, access control, rotation, and audit logging for sensitive credentials.

### Deployment Section

**Q16**: 3 replicas ensure high availability across failure domains and allow rolling updates without downtime.

**Q19**: Liveness probes determine if a container needs to be restarted. Readiness probes determine if a container should receive traffic.

**Q25**: Pod anti-affinity spreads pods across nodes/zones, preventing a single failure from taking down all replicas.

---

## Scoring

- **90-100% (27-30)**: Excellent - Ready for production deployment
- **80-89% (24-26)**: Good - Review missed topics before deployment
- **70-79% (21-23)**: Fair - Additional study strongly recommended
- **Below 70%**: Review modules 10-11 thoroughly

## Certification

Completing this quiz with 80%+ along with all previous quizzes qualifies you for:
- HelixAgent Developer Certification
- Access to advanced deployment guides
- Community contributor status

## Next Steps

After passing this quiz:
1. Complete Lab 5: Production Deployment
2. Review the full DEVELOPER_GUIDE.md
3. Set up a staging environment
4. Plan your production deployment

---

## Additional Resources

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [OWASP LLM Top 10](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
- [HelixAgent Security Documentation](../../security/)
- [Production Deployment Checklist](../../guides/production-checklist.md)
