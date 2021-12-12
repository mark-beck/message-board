# pull official base image
FROM node:lts-alpine3.14 as builder

# set working directory
WORKDIR /app

# add `/app/node_modules/.bin` to $PATH
ENV PATH /app/node_modules/.bin:$PATH

# install app dependencies
COPY package.json ./
COPY package-lock.json ./
RUN npm install
RUN npm install react-scripts -g

# add app
COPY . ./

RUN npm run build

# FROM node:lts-alpine3.14
FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
RUN rm -rf /etc/nginx/conf.d
COPY conf /etc/nginx

EXPOSE 80

# start app
CMD ["nginx", "-g", "daemon off;"]
