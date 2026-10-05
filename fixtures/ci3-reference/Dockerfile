FROM php@sha256:620a6b9f4d4feef2210026172570465e9d0c1de79766418d3affd09190a7fda5
# Offline build: compiler, PHP source and OS libraries belong to the pinned base.
RUN docker-php-ext-install -j1 mysqli pdo_mysql
WORKDIR /app
COPY dependencies/unzip_6.0-26+deb11u1_amd64.deb /tmp/unzip.deb
RUN dpkg -i /tmp/unzip.deb && rm /tmp/unzip.deb
COPY . /app
RUN php dependencies/composer-2.7.7.phar install --no-dev --no-scripts --no-plugins --no-interaction --no-progress --prefer-dist \
    && php dependencies/composer-2.7.7.phar check-platform-reqs --no-dev \
    && find application -name '*.php' -exec php -l {} \; \
    && rm -rf /root/.composer /root/.cache
USER 1000:1000
EXPOSE 8080
CMD ["php", "-d", "opcache.enable=0", "-S", "0.0.0.0:8080", "router.php"]
