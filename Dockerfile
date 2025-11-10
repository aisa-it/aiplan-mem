FROM scratch

WORKDIR /app
COPY bin/app /app/app

WORKDIR /data
ENV AIPLAN_MEM_PATH=/data/aiplanmem.db

CMD ["/app/app"]
