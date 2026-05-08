# Deploy Assets

This directory contains the production deployment skeleton for `article-sentinel`.

## Layout

- `nginx/article-sentinel.conf`: same-domain reverse-proxy and SPA template
- `systemd/article-sentinel*.service`: long-running services and migrate template
- `scripts/install-release.sh`: verify tarball checksum and extract a release into `/srv/article-sentinel/releases/<version>`
- `scripts/activate-release.sh`: run the version-bound migration, switch `current`, restart services, and run smoke checks
- `scripts/rollback-release.sh`: switch `current` back to a known release and run smoke checks again

## Expected host paths

- Release root: `/srv/article-sentinel`
- Runtime config: `/etc/article-sentinel/config.yaml`
- Runtime env overrides: `/etc/article-sentinel/article-sentinel.env`

## Typical flow

```bash
sudo bash deploy/scripts/install-release.sh \
  --tarball /tmp/article-sentinel_<version>_linux_amd64.tar.gz \
  --version <version> \
  --checksum /tmp/article-sentinel_<version>_linux_amd64.tar.gz.sha256

sudo bash deploy/scripts/activate-release.sh --version <version>

sudo bash deploy/scripts/rollback-release.sh
```

## Environment overrides

- `ARTICLE_SENTINEL_DEPLOY_ROOT`: override `/srv/article-sentinel`
- `ARTICLE_SENTINEL_SMOKE_BASE_URL`: override the smoke base URL, default `http://127.0.0.1`

`activate-release.sh` updates `previous` only after the new release passes smoke checks so rollback continues to point at the last known-good version.
