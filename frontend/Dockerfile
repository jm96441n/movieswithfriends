FROM node:18-alpine

WORKDIR /app

COPY frontend/package.json /app/
COPY frontend/package-lock.json /app/
RUN npm install

COPY frontend/tailwind.config.js /app/
COPY frontend/public/ /app/public
COPY frontend/src/ /app/src


CMD ["npm", "start"]
