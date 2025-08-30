version: '3.9'

services:
    postgres:
        image: postgres:15
        environment:
            POSTGRES_DB: ${POSTGRES_DB}
            POSTGRES_USER: ${POSTGRES_USER}
            POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
        volumes:
            - postgres_data:/var/lib/postgresql/data
            - ./database/init.sql:/docker-entrypoint-initdb.d/init.sql
        healthcheck:
            test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
            interval: 5s
            timeout: 5s
            retries: 10
        networks:
            - finance_network
        restart: unless-stopped

    make-data-service:
        build:
            context: .
            dockerfile: make-data-service/Dockerfile
        environment:
            ALPHA_API: ${ALPHA_API}
            ALLOWED_ORIGINS: ${ALLOWED_ORIGINS}
        networks:
            - finance_network
        expose:
            - "8002"
            - "8005"
        restart: unless-stopped

    auth:
        build: ./auth-service
        environment:
            JWT_SECRET: ${JWT_SECRET}
            DB_HOST: postgres
            DB_USER: ${POSTGRES_USER}
            DB_PASSWORD: ${POSTGRES_PASSWORD}
            DB_NAME: ${POSTGRES_DB}
            ALLOWED_ORIGINS: ${ALLOWED_ORIGINS}
        depends_on:
            postgres:
                condition: service_healthy
        restart: unless-stopped
        networks:
            - finance_network
        expose:
            - "8001"

    portfolio:
        build:
            context: .
            dockerfile: portfolio-service/Dockerfile
        environment:
            DB_HOST: postgres
            DB_USER: ${POSTGRES_USER}
            DB_PASSWORD: ${POSTGRES_PASSWORD}
            DB_NAME: ${POSTGRES_DB}
            AUTH_SERVICE_URL: http://auth:8001
            MARKET_SERVICE_URL: http://make-data-service:8002
            MARKET_GRPC_URL: make-data-service:8005
            ALLOWED_ORIGINS: ${ALLOWED_ORIGINS}
        depends_on:
            postgres:
                condition: service_healthy
            auth:
                condition: service_started
        networks:
            - finance_network
        expose:
            - "8003"
        restart: unless-stopped

    frontend:
        build:
            context: ./frontend
            dockerfile: Dockerfile
        container_name: finance_frontend
        depends_on:
            - auth
            - portfolio
            - make-data-service
        networks:
            - finance_network
        expose:
            - "80"
        restart: unless-stopped

    nginx:
        build:
            context: ./nginx
            dockerfile: Dockerfile
        container_name: finance_nginx
        ports:
            - "80:80"
            - "443:443"  # For future SSL
        depends_on:
            - frontend
            - auth
            - portfolio
            - make-data-service
        networks:
            - finance_network
        restart: unless-stopped
        volumes:
            - ./ssl:/etc/nginx/ssl  # For SSL certificates

networks:
    finance_network:
        driver: bridge

volumes:
    postgres_data: