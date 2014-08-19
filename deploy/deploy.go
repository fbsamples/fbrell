// Command deploy augments the deploy process for rell.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/facebookgo/stackerr"
	"github.com/samalba/dockerclient"
)

const (
	prodNginxConfFile       = "prod.conf"
	infoCheckMaxWait        = time.Minute
	infoCheckSleep          = 25 * time.Millisecond
	nginxConfDir            = "/etc/nginx/server"
	nginxPidFile            = "/run/nginx.pid"
	redisContainerLink      = "redis:redis"
	redisContainerName      = "redis"
	redisDataBind           = "/var/lib/redis:/data"
	redisImage              = "daaku/redis"
	rellContainerNamePrefix = "rell"
	rellEnvFile             = "/etc/conf.d/rell"
	rellImage               = "daaku/rell"
	rellPort                = 43600
	rellUser                = "15151"
)

type Deploy struct {
	DockerURL    string
	ServerSuffix string

	client     *dockerclient.DockerClient
	clientErr  error
	clientOnce sync.Once
}

func (d *Deploy) docker() (*dockerclient.DockerClient, error) {
	d.clientOnce.Do(func() {
		d.client, d.clientErr = dockerclient.NewDockerClient(d.DockerURL, nil)
	})
	return d.client, stackerr.Wrap(d.clientErr)
}

func (d *Deploy) startRedis() error {
	docker, err := d.docker()
	if err != nil {
		return err
	}

	ci, err := docker.InspectContainer(redisContainerName)

	// already running, we're set
	if err == nil && ci.State.Running {
		return nil
	}

	// some unknown error, bail
	if err != nil && err != dockerclient.ErrNotFound {
		return stackerr.Wrap(err)
	}

	// exists but not running, remove it and start fresh
	if err == nil {
		if err := docker.RemoveContainer(redisContainerName); err != nil {
			return stackerr.Wrap(err)
		}
	}

	// need to create the container and start it
	containerConfig := dockerclient.ContainerConfig{
		Image: redisImage,
	}
	id, err := docker.CreateContainer(&containerConfig, redisContainerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	hostConfig := dockerclient.HostConfig{
		Binds: []string{redisDataBind},
	}
	err = docker.StartContainer(id, &hostConfig)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) startRell(tag string) error {
	docker, err := d.docker()
	if err != nil {
		return err
	}

	containerName := containerNameForTag(tag)
	ci, err := docker.InspectContainer(containerName)

	// already running, we're set
	if err == nil && ci.State.Running {
		return nil
	}

	// some unknown error, bail
	if err != nil && err != dockerclient.ErrNotFound {
		return stackerr.Wrap(err)
	}

	// exists but not running, remove it and start fresh
	if err == nil {
		if err := docker.RemoveContainer(containerName); err != nil {
			return stackerr.Wrap(err)
		}
	}

	// build our env from the config file
	env, err := d.rellEnv()
	if err != nil {
		return err
	}

	// need to create the container and start it
	containerConfig := dockerclient.ContainerConfig{
		User:  rellUser,
		Image: fmt.Sprintf("%s:%s", rellImage, tag),
		Env:   env,
	}
	id, err := docker.CreateContainer(&containerConfig, containerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	hostConfig := dockerclient.HostConfig{
		Links: []string{redisContainerLink},
	}
	err = docker.StartContainer(id, &hostConfig)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) rellEnv() ([]string, error) {
	contents, err := ioutil.ReadFile(rellEnvFile)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}

	lines := bytes.Split(contents, []byte("\n"))
	env := make([]string, 0, len(lines))
	for _, l := range lines {
		str := strings.TrimSpace(string(l))
		if str == "" {
			continue
		}
		env = append(env, str)
	}
	return env, nil
}

func (d *Deploy) genTagNginxConf(tag string) error {
	docker, err := d.docker()
	if err != nil {
		return err
	}

	containerName := containerNameForTag(tag)
	ci, err := docker.InspectContainer(containerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	filename := filepath.Join(nginxConfDir, containerName+".conf")
	f, err := os.Create(filename)
	if err != nil {
		return stackerr.Wrap(err)
	}

	data := struct {
		BackendName string
		ServerName  string
		IpAddress   string
		Port        int
	}{
		BackendName: containerName,
		ServerName:  fmt.Sprintf("%s.%s", tag, d.ServerSuffix),
		IpAddress:   ci.NetworkSettings.IpAddress,
		Port:        rellPort,
	}
	if err = tagNginxConf.Execute(f, data); err != nil {
		f.Close()
		os.Remove(filename)
		return stackerr.Wrap(err)
	}

	if err := f.Close(); err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) genProdNginxConf(tag string) error {
	containerName := containerNameForTag(tag)

	filename := filepath.Join(nginxConfDir, prodNginxConfFile)
	f, err := os.Create(filename)
	if err != nil {
		return stackerr.Wrap(err)
	}

	data := struct {
		ServerSuffix string
		BackendName  string
	}{
		ServerSuffix: d.ServerSuffix,
		BackendName:  containerName,
	}
	if err = prodNginxConf.Execute(f, data); err != nil {
		f.Close()
		os.Remove(filename)
		return stackerr.Wrap(err)
	}

	if err := f.Close(); err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) hupNginx() error {
	pidStr, err := ioutil.ReadFile(nginxPidFile)
	if err != nil {
		return stackerr.Wrap(err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidStr)))
	if err != nil {
		return stackerr.Wrap(err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return stackerr.Wrap(err)
	}

	err = process.Signal(syscall.SIGHUP)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) infoCheck(tag string) error {
	docker, err := d.docker()
	if err != nil {
		return err
	}

	containerName := containerNameForTag(tag)
	ci, err := docker.InspectContainer(containerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	u := fmt.Sprintf("http://%s:%d/info/", ci.NetworkSettings.IpAddress, rellPort)
	until := time.Now().Add(infoCheckMaxWait)
	for {
		res, err := http.Head(u)
		if err != nil {
			if time.Now().After(until) {
				return stackerr.Wrap(err)
			}
			time.Sleep(infoCheckSleep)
			continue
		}
		res.Body.Close()
		break
	}

	return nil
}

func (d *Deploy) DeployTag(tag string) error {
	if err := d.startRedis(); err != nil {
		return err
	}

	if err := d.startRell(tag); err != nil {
		return err
	}

	if err := d.genTagNginxConf(tag); err != nil {
		return err
	}

	if err := d.genProdNginxConf(tag); err != nil {
		return err
	}

	if err := d.hupNginx(); err != nil {
		return err
	}

	if err := d.infoCheck(tag); err != nil {
		return err
	}

	return nil
}

func containerNameForTag(tag string) string {
	return fmt.Sprintf("%s-%s", rellContainerNamePrefix, tag)
}

// getenv is like os.Getenv, but returns def if the variable is empty.
func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func main() {
	d := Deploy{
		DockerURL:    getenv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		ServerSuffix: getenv("SERVER_SUFFIX", "minetti.fbrell.com"),
	}
	err := d.DeployTag(getenv("TAG", "latest"))
	if err != nil {
		log.Fatal(err)
	}
}

var tagNginxConf = template.Must(template.New("tag").Parse(
	`upstream {{.BackendName}} {
  server               {{.IpAddress}}:{{.Port}};
}
server {
  listen               [::]:80;
  server_name          {{.ServerName}};
  charset              utf-8;
  access_log           off;

  location / {
    proxy_pass         http://{{.BackendName}};
    proxy_set_header   X-Forwarded-For         $remote_addr;
    proxy_set_header   X-Forwarded-Proto       http;
    proxy_set_header   X-Forwarded-Host        $host;
  }
}
server {
  listen               [::]:443;
  server_name          {{.ServerName}};
  ssl                  on;
  ssl_certificate      /etc/nginx/cert/star-minetti-cert.pem;
  ssl_certificate_key  /etc/nginx/cert/star-minetti-key.pem;
  ssl_prefer_server_ciphers on;
  ssl_ciphers 'kEECDH+ECDSA+AES128 kEECDH+ECDSA+AES256 kEECDH+AES128 kEECDH+AES256 kEDH+AES128 kEDH+AES256 DES-CBC3-SHA +SHA !aNULL !eNULL !LOW !MD5 !EXP !DSS !PSK !SRP !kECDH !CAMELLIA !RC4 !SEED';
  ssl_session_cache    shared:SSL:10m;
  ssl_session_timeout  10m;
  keepalive_timeout    70;
  ssl_buffer_size      1400;
  spdy_headers_comp    0;

  charset              utf-8;
  access_log           off;

  location / {
    proxy_pass         http://{{.BackendName}};
    proxy_set_header   X-Forwarded-For         $remote_addr;
    proxy_set_header   X-Forwarded-Proto       https;
    proxy_set_header   X-Forwarded-Host        $host;
  }
}
`))

var prodNginxConf = template.Must(template.New("prod").Parse(
	`server {
  listen               [::]:80;
  server_name          {{.ServerSuffix}};

  location / {
    rewrite (.*) http://www.{{.ServerSuffix}}$1 permanent;
  }
}
server {
  listen               [::]:443;
  server_name          {{.ServerSuffix}};
  ssl                  on;
  ssl_certificate      /etc/nginx/cert/star-minetti-cert.pem;
  ssl_certificate_key  /etc/nginx/cert/star-minetti-key.pem;
  ssl_prefer_server_ciphers on;
  ssl_ciphers 'kEECDH+ECDSA+AES128 kEECDH+ECDSA+AES256 kEECDH+AES128 kEECDH+AES256 kEDH+AES128 kEDH+AES256 DES-CBC3-SHA +SHA !aNULL !eNULL !LOW !MD5 !EXP !DSS !PSK !SRP !kECDH !CAMELLIA !RC4 !SEED';

  location / {
    rewrite (.*) https://www.{{.ServerSuffix}}$1 permanent;
  }
}
server {
  listen               [::]:80;
  server_name          www.{{.ServerSuffix}};
  charset              utf-8;
  access_log           off;

  location / {
    proxy_pass         http://{{.BackendName}};
    proxy_set_header   X-Forwarded-For         $remote_addr;
    proxy_set_header   X-Forwarded-Proto       http;
    proxy_set_header   X-Forwarded-Host        $host;
  }
}
server {
  listen               [::]:443 ssl spdy ipv6only=off;
  server_name          www.{{.ServerSuffix}};
  ssl                  on;
  ssl_certificate      /etc/nginx/cert/star-minetti-cert.pem;
  ssl_certificate_key  /etc/nginx/cert/star-minetti-key.pem;
  ssl_prefer_server_ciphers on;
  ssl_ciphers 'kEECDH+ECDSA+AES128 kEECDH+ECDSA+AES256 kEECDH+AES128 kEECDH+AES256 kEDH+AES128 kEDH+AES256 DES-CBC3-SHA +SHA !aNULL !eNULL !LOW !MD5 !EXP !DSS !PSK !SRP !kECDH !CAMELLIA !RC4 !SEED';
  ssl_session_cache    shared:SSL:10m;
  ssl_session_timeout  10m;
  keepalive_timeout    70;
  ssl_buffer_size      1400;
  spdy_headers_comp    0;

  charset              utf-8;
  access_log           off;

  location / {
    proxy_pass         http://{{.BackendName}};
    proxy_set_header   X-Forwarded-For         $remote_addr;
    proxy_set_header   X-Forwarded-Proto       https;
    proxy_set_header   X-Forwarded-Host        $host;
  }
}
`))
