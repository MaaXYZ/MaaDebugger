import { addCollection } from "@iconify/vue";
import { icons as LucideIcons } from "@iconify-json/lucide";
import { icons as SimpleIcons } from "@iconify-json/simple-icons";

export default function addIcons() {
  addLucideIcons();
  addSimpleIcons();
}

function addLucideIcons() {
  addCollection(LucideIcons);
}

function addSimpleIcons() {
  addCollection(SimpleIcons);
}
