/interface bridge
add name=hs_bridge

/ip hotspot profile
set [ find default=yes ] login-by=mac mac-auth-password=macauth radius-interim-update=1m use-radius=yes

/ip hotspot user profile
set [ find default=yes ] keepalive-timeout=30s

/ip pool
add name=hs-dhcp ranges=10.12.12.10-10.12.12.230

/ip dhcp-server
add address-pool=hs-dhcp interface=hs_bridge lease-time=1h name=server1

/ip hotspot
add address-pool=hs-dhcp addresses-per-mac=unlimited disabled=no idle-timeout=1m interface=hs_bridge keepalive-timeout=30s login-timeout=5s name=server1

/interface bridge port
remove [ find bridge=bridge comment=defconf interface=wifi1 internal-path-cost=10 path-cost=10 ]
add bridge=hs_bridge comment=defconf interface=wifi1 internal-path-cost=10 path-cost=10

/ip address
add address="10.12.12.1/24" interface=hs_bridge network=10.12.12.0

/ip dhcp-server network
add address="10.12.12.0/24" gateway=10.12.12.1

/radius
add address=192.168.88.235 service=hotspot

/ip hotspot walled-garden
add comment="place hotspot rules here" disabled=yes
add comment="captive portal" dst-host=192.168.88.235 dst-port=8989 server=server1
add dst-host=:^gitea.example.com
/ip hotspot walled-garden ip
add action=accept disabled=no dst-address=192.168.88.235 !dst-address-list dst-port=8989 protocol=tcp server=server1 \
    !src-address !src-address-list
