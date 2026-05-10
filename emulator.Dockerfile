FROM eclipse-temurin:21-jre-jammy

RUN apt-get update && apt-get install -y curl gnupg && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | \
    gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | \
    tee /etc/apt/sources.list.d/google-cloud-sdk.list && \
    apt-get update && \
    apt-get install -y google-cloud-sdk-firestore-emulator && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

EXPOSE 8080
CMD ["gcloud", "emulators", "firestore", "start", "--host-port=0.0.0.0:8080", "--project=demo-itsumo"]
