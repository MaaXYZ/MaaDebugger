import { computed } from "vue";
import { useResourceStore } from "@/stores/resource";
import { loadResource } from "@/api/http";

export function useResourceControl() {
  const toast = useToast();
  const resourceStore = useResourceStore();
  const enabledPaths = computed(() => resourceStore.getEnabledPaths());

  async function tryLoadResource(): Promise<{
    success: boolean;
    msg?: string;
  }> {
    const paths = enabledPaths.value;
    if (paths.length === 0) return { success: false, msg: "No enabled paths" };

    try {
      const result = await loadResource(paths);
      if (!result.succeed) {
        console.error("[Resource] Load failed:", result.msg);
        return { success: false, msg: result.msg };
      }
      return { success: true };
    } catch (err) {
      console.error("[Resource] Load failed:", err);
      return { success: false, msg: String(err) };
    }
  }

  async function onLoadResource() {
    const { success, msg } = await tryLoadResource();
    if (!success) {
      console.error("[Resource] Load failed:", msg);
      toast.add({
        id: "res-toast",
        title: "Resource Load Failed",
        description: msg || "Unknown error",
        icon: "i-lucide-circle-x",
        color: "error",
      });
    } else {
      toast.add({
        id: "res-toast",
        title: "Resource Loaded",
        icon: "i-lucide-check-circle",
        color: "success",
      });
    }
  }

  return {
    enabledPaths,
    tryLoadResource,
    onLoadResource,
  };
}
