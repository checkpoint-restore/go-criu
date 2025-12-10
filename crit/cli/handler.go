package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/checkpoint-restore/go-criu/v8/crit"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/apparmor"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/autofs"
	binfmt_misc "github.com/checkpoint-restore/go-criu/v8/crit/images/binfmt-misc"
	bpfmap_data "github.com/checkpoint-restore/go-criu/v8/crit/images/bpfmap-data"
	bpfmap_file "github.com/checkpoint-restore/go-criu/v8/crit/images/bpfmap-file"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/cgroup"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/cpuinfo"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/creds"
	criu_core "github.com/checkpoint-restore/go-criu/v8/crit/images/criu-core"
	criu_sa "github.com/checkpoint-restore/go-criu/v8/crit/images/criu-sa"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/eventfd"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/eventpoll"
	ext_file "github.com/checkpoint-restore/go-criu/v8/crit/images/ext-file"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fdinfo"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fh"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fifo"
	file_lock "github.com/checkpoint-restore/go-criu/v8/crit/images/file-lock"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fs"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/fsnotify"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/inventory"
	ipc_msg "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-msg"
	ipc_sem "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-sem"
	ipc_shm "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-shm"
	ipc_var "github.com/checkpoint-restore/go-criu/v8/crit/images/ipc-var"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/memfd"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mm"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/mnt"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/netdev"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/ns"
	packet_sock "github.com/checkpoint-restore/go-criu/v8/crit/images/packet-sock"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pidns"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pipe"
	pipe_data "github.com/checkpoint-restore/go-criu/v8/crit/images/pipe-data"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/pstree"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/regfile"
	remap_file_path "github.com/checkpoint-restore/go-criu/v8/crit/images/remap-file-path"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/rlimit"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/seccomp"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/signalfd"
	sk_inet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-inet"
	sk_netlink "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-netlink"
	sk_packet "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-packet"
	sk_unix "github.com/checkpoint-restore/go-criu/v8/crit/images/sk-unix"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/stats"
	tcp_stream "github.com/checkpoint-restore/go-criu/v8/crit/images/tcp-stream"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/timens"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/timer"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/timerfd"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/tty"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/tun"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/userns"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/utsns"
	"github.com/checkpoint-restore/go-criu/v8/crit/images/vma"
	"google.golang.org/protobuf/proto"
)

func GetEntryTypeFromImg(imgFile *os.File) (proto.Message, error) {
	magic, err := crit.ReadMagic(imgFile)
	if err != nil {
		return nil, err
	}
	// Seek to the beginning of the file
	_, err = imgFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return protoHandler(magic)
}

func GetEntryTypeFromJSON(jsonFile *os.File) (proto.Message, error) {
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	// Seek to the beginning of the file
	_, err = jsonFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var img map[string]any
	err = json.Unmarshal(jsonData, &img)
	if err != nil {
		return nil, err
	}

	return protoHandler(img["magic"].(string))
}

func protoHandler(magic string) (proto.Message, error) {
	switch magic {
	case "APPARMOR":
		return &apparmor.ApparmorEntry{}, nil
	case "AUTOFS":
		return &autofs.AutofsEntry{}, nil
	case "BINFMT_MISC":
		return &binfmt_misc.BinfmtMiscEntry{}, nil
	case "BPFMAP_DATA":
		return &bpfmap_data.BpfmapDataEntry{}, nil
	case "BPFMAP_FILE":
		return &bpfmap_file.BpfmapFileEntry{}, nil
	case "CGROUP":
		return &cgroup.CgroupEntry{}, nil
	case "CORE":
		return &criu_core.CoreEntry{}, nil
	case "CPUINFO":
		return &cpuinfo.CpuinfoEntry{}, nil
	case "CREDS":
		return &creds.CredsEntry{}, nil
	case "EVENTFD_FILE":
		return &eventfd.EventfdFileEntry{}, nil
	case "EVENTPOLL_FILE":
		return &eventpoll.EventpollFileEntry{}, nil
	case "EVENTPOLL_TFD":
		return &eventpoll.EventpollTfdEntry{}, nil
	case "EXT_FILES":
		return &ext_file.ExtFileEntry{}, nil
	case "FANOTIFY_FILE":
		return &fsnotify.FanotifyFileEntry{}, nil
	case "FANOTIFY_MARK":
		return &fsnotify.FanotifyMarkEntry{}, nil
	case "FDINFO":
		return &fdinfo.FdinfoEntry{}, nil
	case "FIFO":
		return &fifo.FifoEntry{}, nil
	case "FIFO_DATA":
		return &pipe_data.PipeDataEntry{}, nil
	case "FILES":
		return &fdinfo.FileEntry{}, nil
	case "FILE_LOCKS":
		return &file_lock.FileLockEntry{}, nil
	case "FS":
		return &fs.FsEntry{}, nil
	case "IDS":
		return &criu_core.TaskKobjIdsEntry{}, nil
	case "INETSK":
		return &sk_inet.InetSkEntry{}, nil
	case "INOTIFY_FILE":
		return &fsnotify.InotifyFileEntry{}, nil
	case "INOTIFY_WD":
		return &fsnotify.InotifyWdEntry{}, nil
	case "INVENTORY":
		return &inventory.InventoryEntry{}, nil
	case "IPCNS_MSG":
		return &ipc_msg.IpcMsgEntry{}, nil
	case "IPCNS_SEM":
		return &ipc_sem.IpcSemEntry{}, nil
	case "IPCNS_SHM":
		return &ipc_shm.IpcShmEntry{}, nil
	case "IPC_VAR":
		return &ipc_var.IpcVarEntry{}, nil
	case "IRMAP_CACHE":
		return &fh.IrmapCacheEntry{}, nil
	case "ITIMERS":
		return &timer.ItimerEntry{}, nil
	case "MEMFD_INODE":
		return &memfd.MemfdInodeEntry{}, nil
	case "MM":
		return &mm.MmEntry{}, nil
	case "MNTS":
		return &mnt.MntEntry{}, nil
	case "NETDEV":
		return &netdev.NetDeviceEntry{}, nil
	case "NETLINK_SK":
		return &sk_netlink.NetlinkSkEntry{}, nil
	case "NETNS":
		return &netdev.NetnsEntry{}, nil
	case "NS_FILES":
		return &ns.NsFileEntry{}, nil
	case "PACKETSK":
		return &packet_sock.PacketSockEntry{}, nil
	case "PIDNS":
		return &pidns.PidnsEntry{}, nil
	case "PIPES":
		return &pipe.PipeEntry{}, nil
	case "PIPES_DATA":
		return &pipe_data.PipeDataEntry{}, nil
	case "POSIX_TIMERS":
		return &timer.PosixTimerEntry{}, nil
	case "PSTREE":
		return &pstree.PstreeEntry{}, nil
	case "REG_FILES":
		return &regfile.RegFileEntry{}, nil
	case "REMAP_FPATH":
		return &remap_file_path.RemapFilePathEntry{}, nil
	case "RLIMIT":
		return &rlimit.RlimitEntry{}, nil
	case "SECCOMP":
		return &seccomp.SeccompEntry{}, nil
	case "SIGACT":
		return &criu_sa.SaEntry{}, nil
	case "SIGNALFD":
		return &signalfd.SignalfdEntry{}, nil
	case "SK_QUEUES":
		return &sk_packet.SkPacketEntry{}, nil
	case "STATS":
		return &stats.StatsEntry{}, nil
	case "TCP_STREAM":
		return &tcp_stream.TcpStreamEntry{}, nil
	case "TIMENS":
		return &timens.TimensEntry{}, nil
	case "TIMERFD":
		return &timerfd.TimerfdEntry{}, nil
	case "TTY_DATA":
		return &tty.TtyDataEntry{}, nil
	case "TTY_FILES":
		return &tty.TtyFileEntry{}, nil
	case "TTY_INFO":
		return &tty.TtyInfoEntry{}, nil
	case "TUNFILE":
		return &tun.TunfileEntry{}, nil
	case "UNIXSK":
		return &sk_unix.UnixSkEntry{}, nil
	case "USERNS":
		return &userns.UsernsEntry{}, nil
	case "UTSNS":
		return &utsns.UtsnsEntry{}, nil
	case "VMAS":
		return &vma.VmaEntry{}, nil
	/* Pagemap and ghost file have custom handlers
	and cannot use a single proto struct to be
	encoded or decoded. Hence, for these two
	image types, nil is returned. */
	case "PAGEMAP":
		return nil, nil
	case "GHOST_FILE":
		return nil, nil
	}
	return nil, fmt.Errorf("no protobuf binding found for magic 0x%x", magic)
}
