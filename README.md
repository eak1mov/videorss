# VideoRSS

Cloud-friendly rss feed for vkvideo.  
HTTP handle: `http://hostname/vk/{group_name}`.

Technical features:
- In-memory cache.
- Rate-limiter for users.
- Throttler for API.
- Cloud logging.
- Prometheus metrics for cloud monitoring.
- Integration with cloud secrets.

Product features:
- Simple video filter.
- Group whitelist.

For cloud deployment instructions, see `yc/README.md`.
