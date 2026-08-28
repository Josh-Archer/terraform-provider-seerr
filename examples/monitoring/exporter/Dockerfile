FROM python:3.11-alpine

WORKDIR /app

COPY exporter.py /app/exporter.py

ENV EXPORTER_PORT=9850 \
    SEERR_URL=http://localhost:5055 \
    SEERR_API_KEY=""

EXPOSE 9850

USER 65534:65534

ENTRYPOINT ["python", "-u", "/app/exporter.py"]
