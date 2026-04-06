import { ref, onUnmounted, onMounted } from "vue";

export default function useBackoffPolling<
  T extends (...args: any[]) => Promise<any>,
>(
  apiFunc: T,
  enabled: () => boolean,
  interval: number,
  initialArgs: Parameters<T> = [] as any,
) {
  const loading = ref(false);
  let timer: ReturnType<typeof setTimeout> | null = null;

  const fetchData = async (...args: Parameters<T>) => {
    try {
      if (enabled() && !loading.value) {
        loading.value = true;
        await apiFunc(...args);
      }
    } finally {
      loading.value = false;
      setNextTick(interval, ...args);
    }
  };

  function setNextTick(delay: number, ...args: Parameters<T>) {
    stop();
    timer = setTimeout(() => fetchData(...args), delay);
  }

  function stop() {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  }

  // 组件挂载时启动
  onMounted(() => {
    fetchData(...initialArgs);
  });

  // 组件卸载时彻底停止
  onUnmounted(() => {
    enabled = () => false;
    stop();
  });

  return { loading };
}
