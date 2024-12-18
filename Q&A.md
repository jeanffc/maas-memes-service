# MaaS Scalability & Premium Feature – Q&A

## **3.a. How will you handle CI/CD in the context of a live service?**

**Answer:**  
I’m a big fan of keeping things clean and automated. I’d set up a Git-based workflow—maybe GitHub or GitLab—for version control, and plug in something like GitHub Actions or Jenkins to handle the CI/CD pipeline. Every push would automatically run my tests (including unit, integration, and load tests) to ensure the build is solid.

For the deployment, I’d prefer a **blue-green** or **canary** strategy. That way, I spin up a fresh copy of the service alongside the existing one, run health checks, and then cut traffic over only when I’m confident it’s stable. If anything looks off, I can flip back (rollback) instantly.

And of course, **Infrastructure-as-Code** (Terraform/CloudFormation) ties everything in a neat bow, so the environment stays consistent and recoverable.

## **3.b. How will you model and support SLAs? What should operational SLAs be for this service?**

**Answer:**  
I’d aim for a **99.9% uptime**—meaning maybe less than one hour of downtime per month if something goes sideways. For **latency**, I’d try to keep a **p95** under 200ms so most users get a snappy experience. And for the **error rate**, I’d keep it under about 1%.

To back these up, I’d use tools like Prometheus or Datadog for monitoring and Grafana for dashboards. Alerts would go to PagerDuty whenever p95 latency, error rates, or uptime dip below those thresholds. If we breach an SLA, we’d run postmortems to figure out what happened and prevent it next time.

I’d also do the occasional chaos engineering or load testing to simulate real-world conditions and ensure the system remains resilient under pressure.

## **3.c. How do you support geographically diverse clients? As you scale the system out horizontally, how do you continue to keep track of tokens without slowing down the system?**

**Answer:**  
When you have users scattered all over the globe, **multi-region deployments** are the way to go—like launching separate instances in different AWS regions. A load balancer routes each user to the region closest to them, cutting down latency.

Tracking tokens at scale is all about using a **distributed cache**. I’d probably go with a Redis cluster that can replicate across regions. Then each microservice instance stays stateless: it reads and writes token updates to that external store. If performance is critical, I’d use local writes with async replication—so each region handles its own traffic and then syncs globally after the fact.

Meanwhile, I'd keep a close eye on replication lag and network latency. If one region runs hot or experiences issues, the load balancer can fail over or route some traffic to a fallback region. Throw in rate limiting and connection pooling to ensure no single node gets hammered.

## **4. Describe how you would modify the service to now keep track of whether a client is authorized to get AI-generated memes. If a client has this subscription, then they should get AI-memes, and they should get normal memes otherwise. How do you keep track of authorization of a client as we cale the system without slowing down performance?**

**Answer:**  
To handle **Memes AI**, I'd store a simple boolean (`is_premium`) or subscription tier in our `token_balances` table (or a new `subscriptions` table, depending on design). Whenever a user upgrades, we flip `is_premium` to `true`.

Next, I'd tweak the middleware to check this flag. If the user is premium, off they go to the fancy AI meme logic; if not, they get the basic memes. Performance-wise, I'd keep subscription data in a **Redis cache** for lightning-fast lookups instead of hitting the DB every time. When a user’s subscription changes, I’d update Redis right away (and maybe set a short TTL to ensure consistency).

For generating those AI memes, I'd likely have a separate microservice or an external AI API integrated. If it's resource-heavy, I'd queue requests so we don’t choke the main service. Also, I'd consider charging more tokens for AI memes—just to reflect the extra cost or processing time.

At scale, the magic sauce is keeping everything stateless and leveraging caches or distributed DBs. That way, you can spin up (or tear down) as many service instances as you need without worrying about local session data bogging anything down. Everything needed—tokens, subscription flags—is external, meaning horizontal scaling remains straightforward.
