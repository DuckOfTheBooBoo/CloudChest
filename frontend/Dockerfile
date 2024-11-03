# frontend/Dockerfile
# Use the Node.js image to build the Vue.js app
FROM node:18-alpine AS build

# Set working directory and copy files
WORKDIR /app
COPY . .

# Install dependencies and build the app
RUN npm i -g pnpm
RUN pnpm install
RUN pnpm run build

# Use Nginx to serve the static files
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY --from=build /app/nginx.conf /etc/nginx/nginx.conf

# Expose port 80 for Nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
