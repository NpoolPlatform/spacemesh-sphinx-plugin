FROM uhub.service.ucloud.cn/entropypool/ubuntu:jammy

COPY sphinx-plugin-p2 /usr/local/bin/sphinx-plugin-p2

ENV ENV_COIN_LOCAL_API='' \
  ENV_COIN_PUBLIC_API='' \
  ENV_PROXY='' \
  ENV_SYNC_INTERVAL=0 \
  ENV_CONTRACT='' \
  ENV_LOG_DIR=/var/log \
  ENV_LOG_LEVEL=debug \
  ENV_COIN_NET='' \
  ENV_COIN_TYPE=''

ENTRYPOINT ["sphinx-plugin-p2"]
CMD ["help"]
