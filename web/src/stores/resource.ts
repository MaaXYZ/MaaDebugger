import { defineStore } from "pinia";
import { ref, computed } from "vue";

/**
 * 资源路径项
 */
export interface PathItem {
  id: number;
  path: string;
  enabled: boolean;
}

/**
 * 资源 Profile
 */
export interface ResourceProfile {
  id: number;
  name: string;
  paths: PathItem[];
}

/**
 * Resource Store — 管理资源 profiles 和路径列表
 *
 * 通过 persist: true 自动持久化到服务器。
 */
export const useResourceStore = defineStore(
  "resource",
  () => {
    let nextId = 0;
    let nextProfileId = 0;

    const profiles = ref<ResourceProfile[]>([
      { id: nextProfileId++, name: "Default", paths: [] },
    ]);

    const activeProfileId = ref<number>(profiles.value[0]!.id);

    // --- Computed ---
    const activeProfile = computed(() => {
      return (
        profiles.value.find((p) => p.id === activeProfileId.value) ??
        profiles.value[0]!
      );
    });

    const activePaths = computed({
      get: () => activeProfile.value.paths,
      set: (val) => {
        activeProfile.value.paths = val;
      },
    });

    const profileSelectItems = computed(() => {
      return profiles.value.map((p) => ({
        label: p.name,
        value: p.id,
      }));
    });

    // --- Path Actions ---
    function addPath() {
      activePaths.value.push({
        id: nextId++,
        path: "",
        enabled: true,
      });
    }

    function removePath(index: number) {
      activePaths.value.splice(index, 1);
    }

    function reorderPaths(fromIndex: number, toIndex: number) {
      const items = [...activePaths.value];
      const dragged = items.splice(fromIndex, 1)[0];
      if (!dragged) return;
      items.splice(toIndex, 0, dragged);
      activePaths.value = items;
    }

    // --- Profile Actions ---
    function addProfile() {
      const newProfile: ResourceProfile = {
        id: nextProfileId++,
        name: `Profile ${profiles.value.length + 1}`,
        paths: [],
      };
      profiles.value.push(newProfile);
      activeProfileId.value = newProfile.id;
    }

    function deleteProfile() {
      if (profiles.value.length <= 1) return;
      const idx = profiles.value.findIndex(
        (p) => p.id === activeProfileId.value,
      );
      if (idx === -1) return;
      profiles.value.splice(idx, 1);
      activeProfileId.value = profiles.value[0]!.id;
    }

    function renameProfile(newName: string) {
      if (newName.trim()) {
        activeProfile.value.name = newName.trim();
      }
    }

    // --- Public API ---
    function getEnabledPaths(): string[] {
      return activePaths.value
        .filter((item) => item.enabled && item.path)
        .map((item) => item.path);
    }

    function setPaths(newPaths: string[]) {
      activePaths.value = newPaths.map((p) => ({
        id: nextId++,
        path: p,
        enabled: true,
      }));
    }

    /**
     * 确保恢复持久化数据后 nextId/nextProfileId 不会冲突
     */
    function syncIds() {
      const maxPathId = profiles.value.reduce((max, p) => {
        return p.paths.reduce((m, path) => Math.max(m, path.id), max);
      }, 0);
      const maxProfileId = profiles.value.reduce(
        (max, p) => Math.max(max, p.id),
        0,
      );
      nextId = maxPathId + 1;
      nextProfileId = maxProfileId + 1;
    }

    return {
      profiles,
      activeProfileId,
      activeProfile,
      activePaths,
      profileSelectItems,
      addPath,
      removePath,
      reorderPaths,
      addProfile,
      deleteProfile,
      renameProfile,
      getEnabledPaths,
      setPaths,
      syncIds,
    };
  },
  { persist: true },
);
