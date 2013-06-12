apt_repository "dotcloud-lxc-docker" do
  uri "http://ppa.launchpad.net/dotcloud/lxc-docker/ubuntu/"
  distribution node['lsb']['codename']
  components ["main"]
  keyserver "keyserver.ubuntu.com"
  key "63561DC6"
end

package "lxc-docker"
