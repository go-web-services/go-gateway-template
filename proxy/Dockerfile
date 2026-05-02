FROM nginx:1.25-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf

# Placeholder cert so nginx starts before a real Let’s Encrypt run (replace in prod with host mount of /etc/letsencrypt)
RUN apk add --no-cache openssl && \
    mkdir -p /etc/letsencrypt/live/lomank.com && \
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /etc/letsencrypt/live/lomank.com/privkey.pem \
    -out /etc/letsencrypt/live/lomank.com/fullchain.pem \
    -subj "/CN=lomank.com"
