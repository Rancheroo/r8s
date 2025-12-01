# Bundle Discovery: Comprehensive Resource Inventory

**Date:** November 27, 2025  
**Critical Finding:** Rancher support bundles are COMPLETE cluster state snapshots, not just logs!

## Executive Summary

Previous assumption was **WRONG**: Bundles contain far more than pod logs. They include:
- ✅ **kubectl output for 30+ resource types**
- ✅ **CRD definitions** 
- ✅ **Deployments, Services, Namespaces**
- ✅ **Events, ConfigMaps, RBAC**
- ✅ **System diagnostics**
- ✅ **etcd health metrics**
- ✅ **Networking configuration**
- ✅ **Multiple log sources**

This enables building a **full offline cluster browser** from bundles!

---

## Complete Inventory

### 📦 rke2/kubectl/ (Kubernetes Resources)

#### Core Workloads (8 files, 57KB total)
| File | Size | Contents |
|------|------|----------|
| `pods` | 18.5KB | All pods in cluster |
| `deployments` | 9.9KB | All deployments |
| `daemonsets` | 6.3KB | All daemonsets |
| `replicasets` | 12.6KB | All replicasets |
| `statefulsets` | 718B | All statefulsets |
| `jobs` | 2KB | All jobs |
| `cronjobs` | 19B | All cronjobs |
| `helmcharts` | 1.2KB | Installed Helm charts |

#### Networking (7 files, 23KB total)
| File | Size | Contents |
|------|------|----------|
| `services` | 8.3KB | All services |
| `endpoints` | 5.9KB | Service endpoints |
| `ingress` | 19B | Ingress resources |
| `networkpolicies` | 253B | Network policies |
| `pv` | 19B | Persistent volumes |
| `pvc` | 19B | PV claims |
| `volumeattachments` | 88B | Volume attachments |

#### Configuration & Secrets (2 files, 9.6KB total)
| File | Size | Contents |
|------|------|----------|
| `configmaps` | 9.6KB | All ConfigMaps |
| (secrets not in bundle for security) | - | - |

#### **CRDs & API Extensions** (3 files, 13KB total) 🎯
| File | Size | Contents |
|------|------|----------|
| **`crds`** | **8.4KB** | **CRD definitions!** |
| `apiservices` | 4.7KB | API services |
| `api-resources` | 20.9KB | All API resource types |

#### RBAC (4 files, 50KB total)
| File | Size | Contents |
|------|------|----------|
| `clusterroles` | 11KB | Cluster roles |
| `clusterrolebindings` | 27.5KB | Cluster role bindings |
| `roles` | 1.6KB | Namespace roles |
| `rolebindings` | 8.5KB | Role bindings |

#### Admission Control (2 files, 417B total)
| File | Size | Contents |
|------|------|----------|
| `mutatingwebhookconfigurations` | 185B | Mutating webhooks |
| `validatingwebhookconfigurations` | 232B | Validating webhooks |

#### Cluster Info (4 files, 9KB total)
| File | Size | Contents |
|------|------|----------|
| `nodes` | 929B | Node listing |
| `nodesdescribe` | 244B | Node describe output |
| `namespaces` | 792B | All namespaces |
| `events` | 7KB | **Recent cluster events!** |
| `version` | 88B | Kubernetes version |
| `leases` | 3.5KB | Coordination leases |

---

### 📋 rke2/podlogs/ (Pod Logs)

**60+ log files** across multiple namespaces:

#### calico-system (10 pods × 2 logs = 20 files)
- `calico-kube-controllers-*` (current + previous)
- `calico-node-*` (4 nodes × 2 logs)
- `calico-typha-*` (2 replicas × 2 logs)

#### cattle-monitoring-system (40+ files)
- Prometheus, Grafana, Alertmanager
- pushprox clients/proxies for each component
- kube-state-metrics
- prometheus-adapter
- node-exporter

#### cattle-fleet-system
- `fleet-agent-*`

#### kube-system
- `helm-install-*` jobs
- `etcd-*` control plane

**Note:** Each pod has:
- Current logs: `podname`
- Previous logs: `podname-previous` (crashed containers)

---

### 🖥️ systeminfo/ (System Diagnostics - 30 files)

#### Hardware & Resources
| File | Info |
|------|------|
| `cpuinfo` | CPU details |
| `freem` | Memory usage |
| `dfh` | Disk usage |
| `lsblk` | Block devices |
| `mount` | Mounted filesystems |

#### Operating System
| File | Info |
|------|------|
| `osrelease` | OS version |
| `uname` | Kernel version |
| `hostname` | Node hostname |
| `date` | Timestamp |
| `uptime` | System uptime |

#### Processes & Performance
| File | Info |
|------|------|
| `ps` | Running processes |
| `top` | Top processes |
| `vmstat` | Virtual memory stats |
| `lsof` | Open files |

#### Kernel & Modules
| File | Info |
|------|------|
| `dmesg` | Kernel messages |
| `lsmod` | Loaded modules |
| `sysctla` | Kernel parameters |

#### Network Configuration
| File | Info |
|------|------|
| `etchosts` | /etc/hosts |
| `etcresolvconf` | DNS config |
| `systemd-resolved` | systemd DNS |

#### System Limits
| File | Info |
|------|------|
| `file-max` | Max file handles |
| `file-nr` | Current file handles |
| `ulimit-hard` | Resource limits |

#### Services
| File | Info |
|------|------|
| `service-statusall` | All service statuses |
| `systemd-unit-files` | Systemd units |
| `systemd-units` | Active units |
| `packages-dpkg` | Installed packages |

#### Security
| File | Info |
|------|------|
| `ubuntu-apparmorstatus` | AppArmor profiles |
| `ubuntu-ufw` | Firewall rules |

---

### 📊 etcd/ (etcd Health - 7 files)

| File | Purpose |
|------|---------|
| `memberlist` | etcd cluster members |
| `endpointhealth` | Endpoint health checks |
| `endpointstatus` | Endpoint status |
| `alarmlist` | etcd alarms/warnings |
| `etcd-metrics-127.0.0.1:2381.txt` | Prometheus metrics |
| `findserverdbetcd` | etcd database files |
| `findserverdbsnapshots` | etcd snapshots |

---

### 🌐 networking/ (Network Config - 30+ files)

#### IP Configuration
| File | Info |
|------|------|
| `ipaddrshow` | IPv4 addresses |
| `ipv6addrshow` | IPv6 addresses |
| `iplinkshow` | Network interfaces |

#### Routing
| File | Info |
|------|------|
| `iproute` | IPv4 routing table |
| `ipv6route` | IPv6 routing table |
| `iprule` | IPv4 routing rules |
| `ipv6rule` | IPv6 routing rules |

#### Neighbor Discovery
| File | Info |
|------|------|
| `ipneighbour` | IPv4 ARP table |
| `ipv6neighbour` | IPv6 neighbor cache |

#### Firewall (iptables)
| File | Info |
|------|------|
| `iptables` | IPv4 filter rules |
| `iptablesmangle` | IPv4 mangle table |
| `iptablesnat` | IPv4 NAT rules |
| `iptablessave` | Full iptables dump |
| `ip6tables` | IPv6 filter rules |
| `ip6tablesmangle` | IPv6 mangle table |
| `ip6tablesnat` | IPv6 NAT rules |
| `ip6tablessave` | Full ip6tables dump |
| `nft_ruleset` | nftables rules |

#### Socket Statistics
| File | Info |
|------|------|
| `ss4apn` | IPv4 all sockets |
| `ss6apn` | IPv6 all sockets |
| `ssxapn` | Unix sockets |
| `ssanp` | All protocols |
| `ssitan` | TCP info |
| `sstunlp4` | IPv4 UDP/TCP listening |
| `sstunlp6` | IPv6 UDP/TCP listening |
| `ssuapn` | UDP sockets |
| `sswapn` | Raw sockets |

#### CNI
| File | Info |
|------|------|
| `cni/10-calico.conflist` | Calico CNI config |

#### Other
| File | Info |
|------|------|
| `ethtool` | NIC settings |
| `procnetxfrmstat` | IPsec stats |

---

### 📝 systemlogs/ (System Logs - 4 files)

| File | Contents |
|------|----------|
| `syslog` | Current syslog |
| `syslog.1` | Previous syslog |
| `kern.log` | Kernel log |
| `kern.log.1` | Previous kernel log |

---

### 📝 journald/ (Systemd Journals - 3 files)

| File | Service |
|------|---------|
| `rke2-server` | RKE2 server logs |
| `rancher-system-agent` | Rancher agent logs |
| `cloud-init` | Cloud-init logs |

---

### 🔧 rke2/ (RKE2 Specific - Multiple subdirs)

#### Configuration
| File | Purpose |
|------|---------|
| `version` | RKE2 version |
| `50-rancher.yaml` | Rancher config |
| `rke2-server.service` | Server systemd unit |
| `rke2-agent.service` | Agent systemd unit |

#### Logs
| Directory | Contents |
|-----------|----------|
| `agent-logs/` | Kubelet logs |
| `containerd.log` | Container runtime logs |

#### Certificates
| Directory | Contents |
|-----------|----------|
| `certs/agent/` | Agent certificates |
| `certs/server/` | Server certificates |

#### Container Runtime (crictl/)
| File | Info |
|------|------|
| `pods` | Pod listing from CRI |
| `images` | Container images |
| `info` | containerd info |
| `version` | CRI versions |
| `psa` | Pod security |
| `statsa` | Container stats |
| `imagefsinfo` | Image filesystem |

#### Static Pods (pod-manifests/)
- `kube-apiserver.yaml`
- `kube-controller-manager.yaml`
- `kube-scheduler.yaml`
- `kube-proxy.yaml`
- `etcd.yaml`
- `cloud-controller-manager.yaml`

---

## Impact on r8s Development

### What This Enables

#### ✅ Already Implemented
1. **Pod viewing** from bundle
2. **Log viewing** from bundle (current + previous)

#### 🎯 Now Possible (New Discoveries)
1. **CRD browser** - Full CRD support from bundles
2. **Deployment viewer** - Real deployments, not mocks
3. **Service browser** - Real services from bundle
4. **Namespace navigation** - Real namespaces
5. **Event timeline** - Cluster events for diagnostics
6. **Node viewer** - Node status and describe
7. **etcd health dashboard** - etcd cluster health
8. **System diagnostics** - OS, hardware, network info
9. **RBAC viewer** - Roles and bindings
10. **ConfigMap browser** - Configuration data

#### 🚀 Future Enhancements
1. **Event correlation** - Match events to pod issues
2. **Multi-log viewer** - journald + syslog + pod logs
3. **Network diagnostics** - Visualize networking
4. **Resource graphs** - CPU/memory trends from metrics
5. **Security audit** - RBAC, AppArmor, firewall analysis

---

## Bundle Format Analysis

### File Types

99% of files are **plain text** in these formats:

1. **kubectl output** (wide format)
   - Tab or space-separated columns
   - Header row with column names
   - Example: `NAME  READY  STATUS  RESTARTS  AGE`

2. **YAML** (manifests)
   - Static pod definitions
   - RKE2 configuration

3. **JSON** (potential in some kubectl outputs)
   - Could be `-o json` format

4. **Raw logs**
   - Line-oriented text
   - Mixed formats (syslog, journald, application)

5. **Command output**
   - Unstructured text from system commands
   - Parse with regex or line-by-line

### Parsing Strategy

For `rke2/kubectl/*` files:
1. **Try kubectl output parser** (whitespace-delimited)
2. **Fallback to line-by-line** reading
3. **Extract into structs** matching rancher.* types

---

## Implementation Priority

### Phase 5B: Core Resources (HIGH)
1. CRDs - 30 min
2. Deployments - 20 min
3. Services - 20 min
4. Namespaces - 15 min

### Phase 5C: Enhancements (MEDIUM)
1. Events - 20 min
2. Nodes - 15 min
3. ConfigMaps - 15 min

### Phase 5D: Advanced (LOW)
1. System diagnostics viewer
2. etcd health dashboard
3. Network visualization
4. Event correlation

---

## Storage Estimate

Typical bundle size: **50-100MB compressed**
- Extracted: 200-500MB
- Mostly text, compresses well
- Largest: logs (pod logs, system logs)
- Smallest: kubectl outputs (highly structured)

Current r8s limit: **100MB** (configurable)
- Adequate for most bundles
- May need tuning for large clusters

---

## Success Criteria

Bundle mode should support:
- ✅ Browse real CRDs from bundle
- ✅ Browse real Deployments from bundle
- ✅ Browse real Services from bundle
- ✅ Browse real Namespaces from bundle
- ✅ View cluster Events
- ✅ View Node status
- ✅ All existing log features

With graceful fallback to mocks if data unavailable.

---

## Next Steps

1. **Examine file formats** - Read sample kubectl output
2. **Add parsers** - Create inventory functions
3. **Extend DataSource** - Add new methods
4. **Wire up TUI** - Connect fetch* functions
5. **Test** - Verify with real bundle
6. **Document** - Update user docs

**Estimated time:** 90-120 minutes for full resource support

---

## Conclusion

This discovery fundamentally changes the scope of bundle support in r8s. Instead of a "log viewer with mock navigation," we can build a **complete offline cluster browser** that provides real insights from support bundles without needing a live cluster connection.

This is a game-changer for support engineers analyzing customer issues!
