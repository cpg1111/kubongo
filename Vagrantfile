# -*- mode: ruby -*-
# # vi: set ft=ruby :

Vagrant.require_version ">= 1.6.0"

$ip = ENV['VAGRANT_IP'] || '172.22.22.28'

Vagrant.configure(2) do |config|
    config.vm.box = ENV['VAGRANT_BOX_OVERRIDE'] || 'bento/ubuntu-14.04'
    config.vm.synced_folder ".", "/vagrant", disabled: true
    config.vm.provider :vmware_fusion do |vw, override|
        override.vm.network :private_network, ip: $ip
    end
    cmd = %Q(
        sudo apt-get update && sudo apt-get install -y curl build-essential git && \
        mkdir -p /home/vagrant/go/src/github.com/cpg1111 && mkdir -p /home/vagrant/go/pkg && mkdir -p /home/vagrant/go/bin && \
        curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz > /home/vagrant/go.tar.gz && \
        sudo tar -C /usr/local -xzf /home/vagrant/go.tar.gz && \
        echo "export GOPATH=\$GOPATH:/home/vagrant/go" >> /home/vagrant/.profile && \
        echo "export PATH=\$PATH:/usr/local/go/bin:/home/vagrant/go/bin/" >> /home/vagrant/.profile && \
        rm /home/vagrant/go.tar.gz && \
        cd /home/vagrant/go/ && \
        source /home/vagrant/.profile && \
        /usr/local/go/bin/go get github.com/tools/godep && /usr/local/go/bin/go install github.com/tools/godep && \
        echo "success!"
    )
    config.vm.provision :shell, inline: cmd
    config.vm.synced_folder '.', '/home/vagrant/go/src/github.com/cpg1111/kubongo', owner: 'vagrant'
end
