// Command deploy augments the deploy process for rell.
package main

import (
	"bytes"
	"flag"
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

	"github.com/facebookgo/flagconfig"
	"github.com/facebookgo/flagenv"
	"github.com/facebookgo/stackerr"
	"github.com/samalba/dockerclient"
)

type Deploy struct {
	DockerURL               string
	ServerSuffix            string
	CertFile                string
	KeyFile                 string
	InfoCheckMaxWait        time.Duration
	InfoCheckSleep          time.Duration
	LastDeployTagFile       string
	NginxConfDir            string
	NginxPidFile            string
	ProdNginxConfFile       string
	RedisContainerLink      string
	RedisContainerName      string
	RedisDataBind           string
	RedisImage              string
	RellContainerNamePrefix string
	RellEnvFile             string
	RellImage               string
	RellPort                int
	RellUser                string
	StopTimeout             time.Duration

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

	ci, err := docker.InspectContainer(d.RedisContainerName)

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
		if err := docker.RemoveContainer(d.RedisContainerName); err != nil {
			return stackerr.Wrap(err)
		}
	}

	// need to create the container and start it
	containerConfig := dockerclient.ContainerConfig{
		Image: d.RedisImage,
	}
	id, err := docker.CreateContainer(&containerConfig, d.RedisContainerName)
	if err != nil {
		if err != dockerclient.ErrNotFound {
			return stackerr.Wrap(err)
		}

		// pull the image
		if err := docker.PullImage(d.RedisImage, "latest"); err != nil {
			return stackerr.Wrap(err)
		}

		id, err = docker.CreateContainer(&containerConfig, d.RedisContainerName)
		if err != nil {
			return stackerr.Wrap(err)
		}
	}

	hostConfig := dockerclient.HostConfig{
		Binds: []string{d.RedisDataBind},
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

	containerName := d.containerNameForTag(tag)
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
		User:  d.RellUser,
		Image: fmt.Sprintf("%s:%s", d.RellImage, tag),
		Env:   env,
	}
	id, err := docker.CreateContainer(&containerConfig, containerName)
	if err != nil {
		if err != dockerclient.ErrNotFound {
			return stackerr.Wrap(err)
		}

		// pull the image
		if err := docker.PullImage(d.RellImage, tag); err != nil {
			return stackerr.Wrap(err)
		}

		id, err = docker.CreateContainer(&containerConfig, containerName)
		if err != nil {
			return stackerr.Wrap(err)
		}
	}

	hostConfig := dockerclient.HostConfig{
		Links: []string{d.RedisContainerLink},
	}
	err = docker.StartContainer(id, &hostConfig)
	if err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) rellEnv() ([]string, error) {
	contents, err := ioutil.ReadFile(d.RellEnvFile)
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

	containerName := d.containerNameForTag(tag)
	ci, err := docker.InspectContainer(containerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	filename := d.containerNginxConfPath(containerName)
	f, err := os.Create(filename)
	if err != nil {
		return stackerr.Wrap(err)
	}

	data := struct {
		BackendName string
		ServerName  string
		IpAddress   string
		Port        int
		CertFile    string
		KeyFile     string
	}{
		BackendName: containerName,
		ServerName:  fmt.Sprintf("%s.%s", tag, d.ServerSuffix),
		IpAddress:   ci.NetworkSettings.IpAddress,
		Port:        d.RellPort,
		CertFile:    d.CertFile,
		KeyFile:     d.KeyFile,
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

func (d *Deploy) switchProd(tag string) error {
	containerName := d.containerNameForTag(tag)

	// rewrite the prod config
	filename := filepath.Join(d.NginxConfDir, d.ProdNginxConfFile)
	f, err := os.Create(filename)
	if err != nil {
		return stackerr.Wrap(err)
	}

	data := struct {
		ServerSuffix string
		BackendName  string
		CertFile     string
		KeyFile      string
	}{
		ServerSuffix: d.ServerSuffix,
		BackendName:  containerName,
		CertFile:     d.CertFile,
		KeyFile:      d.KeyFile,
	}
	if err = prodNginxConf.Execute(f, data); err != nil {
		f.Close()
		os.Remove(filename)
		return stackerr.Wrap(err)
	}

	if err := f.Close(); err != nil {
		return stackerr.Wrap(err)
	}

	// update the last deploy tag file
	os.MkdirAll(filepath.Dir(d.LastDeployTagFile), os.FileMode(0755))
	f, err = os.Create(d.LastDeployTagFile)
	if err != nil {
		return stackerr.Wrap(err)
	}

	if _, err := fmt.Fprint(f, tag); err != nil {
		return stackerr.Wrap(err)
	}

	if err := f.Close(); err != nil {
		return stackerr.Wrap(err)
	}

	return nil
}

func (d *Deploy) hupNginx() error {
	pidStr, err := ioutil.ReadFile(d.NginxPidFile)
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

	containerName := d.containerNameForTag(tag)
	ci, err := docker.InspectContainer(containerName)
	if err != nil {
		return stackerr.Wrap(err)
	}

	u := fmt.Sprintf("http://%s:%d/info/", ci.NetworkSettings.IpAddress, d.RellPort)
	until := time.Now().Add(d.InfoCheckMaxWait)
	for {
		res, err := http.Head(u)
		if err != nil {
			if time.Now().After(until) {
				return stackerr.Wrap(err)
			}
			time.Sleep(d.InfoCheckSleep)
			continue
		}
		res.Body.Close()
		break
	}

	return nil
}

func (d *Deploy) killExcept(tag string) error {
	docker, err := d.docker()
	if err != nil {
		return err
	}

	containers, err := docker.ListContainers(true)
	if err != nil {
		return stackerr.Wrap(err)
	}

	var errs multiError
	for _, c := range containers {
		// check if its our container
		if !strings.HasPrefix(c.Image, d.RellImage+":") {
			continue
		}

		cTag := d.tagFromNames(c.Names)

		// dont kill the production container
		if cTag == tag {
			continue
		}

		// remove nginx config file
		cName := d.containerNameForTag(cTag)
		os.Remove(d.containerNginxConfPath(cName))

		// stop it, ignoring errors
		docker.StopContainer(c.Id, int(d.StopTimeout.Seconds()))

		// remove it
		if err := docker.RemoveContainer(c.Id); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return stackerr.Wrap(errs[0])
	}

	return stackerr.Wrap(errs)
}

func (d *Deploy) DeployTag(tag string, prod bool) error {
	if err := d.startRedis(); err != nil {
		return err
	}

	if err := d.startRell(tag); err != nil {
		return err
	}

	if err := d.genTagNginxConf(tag); err != nil {
		return err
	}

	if err := d.infoCheck(tag); err != nil {
		return err
	}

	if prod {
		if err := d.switchProd(tag); err != nil {
			return err
		}
	}

	if err := d.hupNginx(); err != nil {
		log.Printf("failed to hup nginx: %s\n", err)
	}

	if prod {
		if err := d.killExcept(tag); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deploy) containerNginxConfPath(containerName string) string {
	return filepath.Join(d.NginxConfDir, containerName+".conf")
}

func (d *Deploy) tagFromNames(names []string) string {
	for _, name := range names {
		name = name[1:] // drop the leading /
		if strings.HasPrefix(name, d.RellContainerNamePrefix) {
			return name[len(d.RellContainerNamePrefix):]
		}
	}
	return ""
}

func (d *Deploy) containerNameForTag(tag string) string {
	return fmt.Sprintf("%s%s", d.RellContainerNamePrefix, tag)
}

// getenv is like os.Getenv, but returns def if the variable is empty.
func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

type multiError []error

func (m multiError) Error() string {
	parts := make([]string, 0, len(m))
	for _, e := range m {
		parts = append(parts, e.Error())
	}
	return "multiple errors: " + strings.Join(parts, " | ")
}

func main() {
	log.SetFlags(log.Lshortfile)

	var d Deploy

	flag.StringVar(
		&d.DockerURL,
		"docker_host",
		"unix:///var/run/docker.sock",
		"docker url",
	)
	flag.StringVar(
		&d.ServerSuffix,
		"server_suffix",
		"minetti.fbrell.com",
		"server suffix",
	)
	flag.StringVar(
		&d.CertFile,
		"cert_file",
		"/etc/nginx/cert/rell.pem",
		"ssl cert file",
	)
	flag.StringVar(
		&d.KeyFile,
		"key_file",
		"/etc/nginx/cert/rell.key",
		"ssl key file",
	)

	d.InfoCheckMaxWait = time.Minute
	d.InfoCheckSleep = 25 * time.Millisecond
	d.LastDeployTagFile = "/var/lib/rell/production-tag"
	d.NginxConfDir = "/etc/nginx/server"
	d.NginxPidFile = "/run/nginx.pid"
	d.ProdNginxConfFile = "rell-prod.conf"
	d.RedisContainerLink = "redis:redis"
	d.RedisContainerName = "redis"
	d.RedisDataBind = "/var/lib/redis:/data"
	d.RedisImage = "daaku/redis"
	d.RellContainerNamePrefix = "rell-"
	d.RellEnvFile = "/etc/conf.d/rell"
	d.RellImage = "daaku/rell"
	d.RellPort = 43600
	d.RellUser = "15151"
	d.StopTimeout = 30 * time.Second

	tag := flag.String("tag", defaultTag(d.LastDeployTagFile), "default tag to deploy")
	switchProd := flag.Bool("prod", false, "switch prod to specified tag")

	flagenv.UseUpperCaseFlagNames = true
	flag.Parse()
	flagenv.Parse()
	flagconfig.Parse()

	err := d.DeployTag(*tag, *switchProd)
	if err != nil {
		log.Fatal(err)
	}
}

func defaultTag(lastDeployTagFile string) string {
	b, _ := ioutil.ReadFile(lastDeployTagFile)
	if len(b) > 0 {
		return string(b)
	}
	return "latest"
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
  ssl_certificate      {{.CertFile}};
  ssl_certificate_key  {{.KeyFile}};
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
  ssl_certificate      {{.CertFile}};
  ssl_certificate_key  {{.KeyFile}};
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
  ssl_certificate      {{.CertFile}};
  ssl_certificate_key  {{.KeyFile}};
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
