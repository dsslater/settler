FROM mhart/alpine-node:latest
ENV NPM_CONFIG_LOGLEVEL info

# Create app directory
WORKDIR /Users/davidslater/Documents/Dev/docker

COPY package*.json ./

RUN npm install

# Bundle app source
COPY . .

EXPOSE 3000

CMD [ "node", "webServer.js" ]
