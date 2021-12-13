# pull official base image
FROM node:16.13.1-stretch as builder



# set working directory
WORKDIR /app

# add `/app/node_modules/.bin` to $PATH
ENV PATH /app/node_modules/.bin:$PATH

# install app dependencies
COPY faceflix/package.json ./
COPY faceflix/yarn.lock ./
RUN yarn install
RUN yarn global add react-scripts

# add app
COPY faceflix/. ./

RUN yarn run build

# FROM node:lts-alpine3.14
FROM nginx:stable-alpine

COPY --from=builder /app/build /usr/share/nginx/html
RUN rm -rf /etc/nginx/conf.d
COPY conf /etc/nginx

EXPOSE 80

# start app
CMD ["nginx", "-g", "daemon off;"]
