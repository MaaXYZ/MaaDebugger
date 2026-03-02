import { ref, computed } from "vue";
import { useResourceStore } from "@/stores/resource";
import { loadResource } from "@/api/http";

const toast = useToast();
const resourceStore = useResourceStore();
// --- UI State (not persisted) ---
const showFullCard = ref(true);
const enabledPaths = computed(() => resourceStore.getEnabledPaths());

export default async function onLoadResource() {
  const paths = enabledPaths.value;
  if (paths.length === 0) return;

  try {
    const result = await loadResource(paths);
    if (!result.succeed) {
      console.error("[Resource] Load failed:", result.msg);
      toast.add({
        id: "res-toast",
        title: "Resource Load Failed",
        description: result.msg || "Unknown error",
        icon: "i-lucide-circle-x",
        color: "error",
      });
    } else {
      showFullCard.value = false;
      toast.add({
        id: "res-toast",
        title: "Resource Loaded",
        icon: "i-lucide-check-circle",
        color: "success",
      });
    }
  } catch (err) {
    console.error("[Resource] Load failed:", err);
  }
}
