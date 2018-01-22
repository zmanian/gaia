# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "shell", inline: <<-SHELL
    # and golang 1.9 support
    # official repo doesn't have race detection runtime...
    #add-apt-repository ppa:gophers/archive
    add-apt-repository ppa:longsleep/golang-backports

    # install base requirements
    apt-get update
    apt-get install -y --no-install-recommends make
    apt-get install -y golang-1.9-go
    apt-get install -y language-pack-en

    # cleanup
    apt-get autoremove -y

    # set $PATH
    echo 'export PATH=$PATH:/usr/lib/go-1.9/bin:/home/vagrant/go/bin' >> /home/vagrant/.bash_profile
    echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.bash_profile
    echo 'export LC_ALL=en_US.UTF-8' >> /home/vagrant/.bash_profile

    mkdir -p /home/vagrant/go/bin
    mkdir -p /home/vagrant/go/src/github.com/cosmos
    ln -s /vagrant /home/vagrant/go/src/github.com/cosmos/gaia

    chown -R vagrant:vagrant /home/vagrant/go
    chown vagrant:vagrant /home/vagrant/.bash_profile

    su - vagrant -c 'source /home/vagrant/.bash_profile'
    su - vagrant -c 'cd /home/vagrant/go/src/github.com/cosmos/gaia && make get_vendor_deps'
  SHELL
end
