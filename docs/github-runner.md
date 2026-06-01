# GitHub Runner Deployment Notes

This project can be deployed from a self-hosted runner, but the simplest first setup is still manual Docker Compose.

Recommended future CI/CD flow:

1. Run tests on every pull request.
2. Build a Docker image on every tag.
3. Publish the image to GitHub Container Registry.
4. Let self-hosted instances pull a pinned image tag.

The included CI workflow already runs tests on `main` and publishes Docker images for tags matching `v*`.

For a private single-server deployment, a self-hosted runner can run:

```bash
git pull
docker compose up -d --build
```

Use a runner only on infrastructure you control. Keep `.env`, `config.json`, and `data/` on the server, not in the repository.
