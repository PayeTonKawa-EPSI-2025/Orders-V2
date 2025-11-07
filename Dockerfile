FROM alpine:3.22
WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY build/paye-ton-kawa--orders /app/paye-ton-kawa--orders

RUN chown appuser:appgroup /app/paye-ton-kawa--orders && chmod +x /app/paye-ton-kawa--orders

EXPOSE 8082

USER appuser

ENTRYPOINT ["/app/paye-ton-kawa--orders"]
