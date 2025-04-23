<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close()"/>
      <title>{{ fileStore.req?.name ?? "" }}</title>

      <!--action
          v-if="authStore.user?.perm.modify"
          id="save-button"
          icon="save"
          :label="t('buttons.save')"
          @action="save()"
      /-->

      <action
          icon="preview"
          :label="t('buttons.preview')"
          @action="preview()"
          v-show="isMarkdownFile"
      />
    </header-bar>

    <Breadcrumbs base="/files" noLink/>

    <div
        v-show="(!isJson) && isPreview && isMarkdownFile"
        id="preview-container"
        class="md_preview"
        v-html="previewContent"
    ></div>

    <div class='scrollable'
         v-show="isJson"
    >
      <JsonViewer
          :value="jsonContent"
          :expand-depth="Infinity"
          copyable
          boxed
          expanded
          :show-array-index="false"
      />
    </div>
    <form v-show="!isJson && (!isPreview || !isMarkdownFile)" id="editor"></form>
  </div>
</template>

<script setup lang="ts">
import JsonViewer from "vue-json-viewer";
import 'vue-json-viewer/style.css'
import {files as api} from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";
import ace, {Ace, version as ace_version} from "ace-builds";
import modelist from "ace-builds/src-noconflict/ext-modelist";
import "ace-builds/src-noconflict/ext-language_tools";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import {useAuthStore} from "@/stores/auth";
import {useFileStore} from "@/stores/file";
import {useLayoutStore} from "@/stores/layout";
import {inject, onBeforeUnmount, onMounted, ref, watchEffect} from "vue";
import {useRoute, useRouter} from "vue-router";
import {useI18n} from "vue-i18n";
import {getTheme} from "@/utils/theme";
import {marked} from "marked";

const $showError = inject<IToastError>("$showError")!;

const fileStore = useFileStore();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const {t} = useI18n();

const route = useRoute();
const router = useRouter();

const editor = ref<Ace.Editor | null>(null);

const isPreview = ref(false);
const previewContent = ref("");
const isMarkdownFile =
    fileStore.req?.name.endsWith(".md") ||
    fileStore.req?.name.endsWith(".markdown");

const jsonContent = ref();
const isJson = ref(false)

watchEffect(() => {
  try {
    jsonContent.value = JSON.parse(fileStore.req?.content || "{}");
    isJson.value = true
  } catch {
    //fileContent2.value = fileStore.req?.content || "";
    isJson.value = false
  }
});


onMounted(() => {
  window.addEventListener("keydown", keyEvent);


  const fileContent = fileStore.req?.content || "";

  watchEffect(async () => {
    if (isMarkdownFile && isPreview.value) {
      const new_value = editor.value?.getValue() || "";
      try {
        previewContent.value = await marked(new_value);
      } catch (error) {
        console.error("Failed to convert content to HTML:", error);
        previewContent.value = "";
      }

      const previewContainer = document.getElementById("preview-container");
    }
  });

  ace.config.set(
      "basePath",
      `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
  );

  editor.value = ace.edit("editor", {
    value: fileContent,
    showPrintMargin: false,
    readOnly: fileStore.req?.type === "textImmutable",
    theme: "ace/theme/chrome",
    mode: modelist.getModeForPath(fileStore.req!.name).mode,
    wrap: true,
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    enableSnippets: true,
  });

  if (getTheme() === "dark") {
    editor.value!.setTheme("ace/theme/twilight");
  }

  editor.value.focus();
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);

  editor.value?.destroy();
});

const keyEvent = (event: KeyboardEvent) => {
  if (event.code === "Escape") {
    close();
  }

  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  if (event.key !== "s") {
    return;
  }

  event.preventDefault();
  save();
};


const save = async () => {
  const button = "save";
  buttons.loading("save");

  try {
    await api.put(route.path, editor.value?.getValue());
    editor.value?.session.getUndoManager().markClean();
    buttons.success(button);
  } catch (e: any) {
    buttons.done(button);
    $showError(e);
  }
};
const close = () => {

//  if (!editor.value?.session.getUndoManager().isClean()) {
//    layoutStore.showHover("discardEditorChanges");
//    return;
//  }

  fileStore.updateRequest(null);

  const uri = url.removeLastDir(route.path) + "/";
  router.push({path: uri});
};

const preview = () => {
  isPreview.value = !isPreview.value;
};
</script>

<style>
/* Override only the string color in JsonViewer */
.jv-light .jv-string {
  color: #14588F !important; /* Your custom color */
}

.scrollable {
  overflow-y: scroll;
}

</style>
