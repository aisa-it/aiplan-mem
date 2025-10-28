FROM scratch

WORKDIR /app
COPY bin/app /app/app

ENV AIPLAN_MEM_PATH=/app/aiplanmem.db

CMD ["/app/app"]
