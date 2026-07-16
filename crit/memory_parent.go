package crit

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func canonicalDirectory(path string) (string, os.FileInfo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, fmt.Errorf("resolve checkpoint directory %s: %w", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", nil, fmt.Errorf("resolve checkpoint directory %s: %w", path, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", nil, fmt.Errorf("stat checkpoint directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("checkpoint path %s is not a directory", path)
	}
	return filepath.Clean(resolved), info, nil
}

func (session *memoryReadSession) ensureParent() error {
	if session.parent != nil {
		return nil
	}
	parentPath := filepath.Join(session.layer.directory, "parent")
	parentDirectory, parentIdentity, err := canonicalDirectory(parentPath)
	if err != nil {
		return fmt.Errorf("open parent for %s: %w", session.layer.pagemapName, err)
	}
	if _, exists := session.visited[parentDirectory]; exists {
		return fmt.Errorf("parent checkpoint cycle detected at %s", parentDirectory)
	}
	for _, identity := range session.identities {
		if os.SameFile(identity, parentIdentity) {
			return fmt.Errorf("parent checkpoint cycle detected at %s", parentDirectory)
		}
	}

	parentLayer, err := loadMemoryLayer(parentDirectory, session.layer.pagemapName, session.layer.pageSize)
	if err != nil {
		return fmt.Errorf("load parent for %s: %w", session.layer.pagemapName, err)
	}
	if parentLayer.hasInventory &&
		parentLayer.compression.ImageVersion == crtoolsImagesV1_2 &&
		(!session.layer.hasInventory ||
			session.layer.compression.ImageVersion < crtoolsImagesV1_2) {
		childVersion := "missing"
		if session.layer.hasInventory {
			childVersion = strconv.FormatUint(
				uint64(session.layer.compression.ImageVersion),
				10,
			)
		}
		return fmt.Errorf(
			"%s image version %s does not propagate V1.2 from its parent",
			session.layer.pagemapName,
			childVersion,
		)
	}
	session.visited[parentDirectory] = struct{}{}
	session.parent = &memoryReadSession{
		layer:      parentLayer,
		visited:    session.visited,
		identities: append(session.identities, parentIdentity),
	}
	return nil
}

func (session *memoryReadSession) readParentPageInto(vaddr uint64, page []byte) (bool, error) {
	if err := session.ensureParent(); err != nil {
		return false, err
	}
	found, err := session.parent.readPageInto(vaddr, page)
	if err != nil {
		return false, err
	}
	if !found {
		return false, fmt.Errorf("parent checkpoint has no page at address %#x", vaddr)
	}
	return true, nil
}
