# ---------------------------------------------------------------------------
# Network + volume
# ---------------------------------------------------------------------------

resource "docker_network" "motd" {
  name = "motd-network"
}

resource "docker_volume" "motd_data" {
  name = "motd-data"
}

# ---------------------------------------------------------------------------
# Secrets
# ---------------------------------------------------------------------------

resource "random_password" "db_password" {
  length  = 20
  special = false
}

resource "random_password" "api_token" {
  length  = 16
  special = false
}

# ---------------------------------------------------------------------------
# Database container
# ---------------------------------------------------------------------------

resource "docker_image" "mysql" {
  name = "mysql:8.4"
}

resource "docker_container" "mysql" {
  name    = "motd-mysql"
  image   = docker_image.mysql.image_id
  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.motd.name
  }

  env = [
    "MYSQL_ROOT_PASSWORD=${random_password.db_password.result}",
    "MYSQL_DATABASE=motd",
  ]

  mounts {
    target = "/var/lib/mysql"
    source = docker_volume.motd_data.name
    type   = "volume"
  }

  mounts {
    target    = "/docker-entrypoint-initdb.d"
    source    = abspath("${path.module}/seed")
    type      = "bind"
    read_only = true
  }

  healthcheck {
    test     = ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-p${random_password.db_password.result}"]
    interval = "5s"
    timeout  = "3s"
    retries  = 20
  }

  wait = true
}

# ---------------------------------------------------------------------------
# Web container
# ---------------------------------------------------------------------------

resource "docker_image" "web" {
  name = "motd-web:latest"

  build {
    context = "${path.module}/app"
  }

  triggers = {
    app_dir_sha1 = sha1(join("", [
      for f in fileset("${path.module}/app", "**") :
      filesha1("${path.module}/app/${f}")
    ]))
  }
}

resource "docker_container" "web" {
  name    = "motd-web"
  image   = docker_image.web.image_id
  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.motd.name
  }

  ports {
    internal = 8080
    external = var.port
  }

  env = [
    "DB_HOST=${docker_container.mysql.name}",
    "DB_PORT=3306",
    "DB_USER=root",
    "DB_PASSWORD=${random_password.db_password.result}",
    "DB_NAME=motd",
    "MOTD_TOKEN=${random_password.api_token.result}",
  ]

  depends_on = [docker_container.mysql]
}

# ---------------------------------------------------------------------------
# Client config file
# ---------------------------------------------------------------------------

resource "local_file" "motd_ini" {
  filename = "${path.module}/config/motd.ini"
  content  = <<-EOT
    token = ${random_password.api_token.result}
    host = localhost
    port = ${var.port}
  EOT
}
