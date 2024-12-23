package persistentvolume

import (
	"strings"

	v1 "k8s.io/api/core/v1"
)

// getAccessModes converts access modes to a readable string
func getAccessModes(modes []v1.PersistentVolumeAccessMode) string {
	if len(modes) == 0 {
		return "N/A"
	}

	modeStrings := make([]string, len(modes))
	for i, mode := range modes {
		switch mode {
		case v1.ReadWriteOnce:
			modeStrings[i] = "RWO"
		case v1.ReadOnlyMany:
			modeStrings[i] = "ROX"
		case v1.ReadWriteMany:
			modeStrings[i] = "RWX"
		default:
			modeStrings[i] = string(mode)
		}
	}

	return strings.Join(modeStrings, ",")
}

// getPVType determines the type of Persistent Volume
func getPVType(pv *v1.PersistentVolume) string {
	switch {
	case pv.Spec.HostPath != nil:
		return "HostPath"
	case pv.Spec.GCEPersistentDisk != nil:
		return "GCEPersistentDisk"
	case pv.Spec.AWSElasticBlockStore != nil:
		return "AWSElasticBlockStore"
	case pv.Spec.NFS != nil:
		return "NFS"
	case pv.Spec.ISCSI != nil:
		return "ISCSI"
	case pv.Spec.Cinder != nil:
		return "Cinder"
	case pv.Spec.CephFS != nil:
		return "CephFS"
	case pv.Spec.FC != nil:
		return "FC"
	case pv.Spec.FlexVolume != nil:
		return "FlexVolume"
	case pv.Spec.AzureFile != nil:
		return "AzureFile"
	case pv.Spec.AzureDisk != nil:
		return "AzureDisk"
	case pv.Spec.Glusterfs != nil:
		return "Glusterfs"
	case pv.Spec.VsphereVolume != nil:
		return "vSphereVolume"
	case pv.Spec.PhotonPersistentDisk != nil:
		return "PhotonPersistentDisk"
	case pv.Spec.PortworxVolume != nil:
		return "PortworxVolume"
	case pv.Spec.ScaleIO != nil:
		return "ScaleIO"
	case pv.Spec.Local != nil:
		return "Local"
	default:
		return "Unknown"
	}
}
