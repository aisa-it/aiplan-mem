FROM scratch

WORKDIR /app
COPY bin/app /app/app
RUN mkdir /app/db

ENV AIPLAN_MEM_PATH=/app/db/aiplanmem.db

CMD ["/app/app"]
