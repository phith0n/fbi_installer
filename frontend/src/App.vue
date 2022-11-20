<template>
  <nav class="navbar bg-light">
    <div class="container">
      <a class="navbar-brand">FBI Remote Game Installer</a>
      <form class="d-flex" role="search">
        <button class="btn btn-primary me-2" type="button"
                @click="config.showConfig = !config.showConfig">Configuration
        </button>
        <div class="col-auto me-2">
          <input class="form-control" type="search" placeholder="Search Games" v-model="config.searchKeyword" @keydown.enter="loadData">
        </div>
        <button class="btn btn-success" type="button" @click="loadData">Search</button>
      </form>
    </div>
  </nav>
  <div class="container mt-4">
    <div v-if="!config.address" class="alert alert-warning" role="alert">
      Your 3DS address is empty, <a href="#" @click="config.showConfig = true" class="text-decoration-none">click
      here</a> to set up the address.
    </div>
    <div v-else class="alert alert-primary" role="alert">
      Your 3DS address is <strong>{{ config.address }}</strong>
    </div>

    <div class="mb-3" v-show="config.showConfig">
      <label for="exampleFormControlInput1" class="form-label">3DS Address</label>
      <input type="text" class="form-control" placeholder="" v-model="config.address">
    </div>

    <hr/>

    <div v-if="config.file == null" :class="['fileupload', dragging ? 'over': '']" @dragenter="dragging = true"
         @dragleave="dragging = false">
      <div class="info" @drag="onFilePut">
        <span class="bi bi-cloud-arrow-up-fill title me-1"></span>
        <span class="title">Drop or click to upload games</span>
        <div class="upload-limit-info">
          <div>extension support: .cia, .3dsx, .cetk, .tik</div>
          <div>maximum file size: 2048 MB</div>
        </div>
      </div>
      <input type="file" @change="onFilePut">
    </div>
    <div v-else class="fileupload-progress d-flex">
      <div class="d-flex w-100 align-items-center justify-content-center">
        <div class="w-75">
          <div class="progress">
            <div class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar"
                 :style="{width: config.progress + '%'}">{{ config.file.name }}
            </div>
          </div>
          <div class="mt-2">
            File Name: {{ config.file.name }} File Size: {{ formatBytes(config.file.size) }}
          </div>
          <div class="mt-2">
            <button type="button" class="btn btn-primary me-2" @click="onUpload">Upload</button>
            <button type="button" class="btn btn-secondary" @click="onCancel">Cancel</button>
          </div>
        </div>
      </div>
    </div>

    <table class="table mt-4">
      <thead>
      <tr>
        <th scope="col">File Name</th>
        <th scope="col">File Size</th>
        <th scope="col">Modify Time</th>
        <th scope="col" class="text-end">Handle</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="game in config.games">
        <th scope="row">{{game.name}}</th>
        <td>{{ formatBytes(game.size) }}</td>
        <td>{{ dayjs.unix(game['mod_time']).format('YYYY-MM-DD HH:mm:ss') }}</td>
        <td class="text-end">
          <button type="button" class="btn btn-outline-danger btn-sm me-2" @click="deleteFile(game.name)">Delete</button>
          <button type="button" class="btn btn-outline-primary btn-sm"
                  :disabled="config.address === ''" @click="sendGame(game.name)">Send to 3DS</button>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import {ref, reactive, onMounted, watch} from "vue";
import axios from 'axios';
import * as dayjs from 'dayjs'
import notie from 'notie'
import "bootstrap-icons/font/bootstrap-icons.scss"
import 'notie/dist/notie.css';

const config = reactive({
  address: "",
  showConfig: false,
  file: null,
  games: [],
  progress: 0,
  searchKeyword: "",
})
const dragging = ref(false);

watch(() => config.address, (current, old) => {
  if (current) {
    window.localStorage.setItem("address", current);
  }
})

const onFilePut = (e) => {
  const files = e.target.files || e.dataTransfer.files;

  if (!files.length) {
    return;
  }

  const file = files[0];
  if (!/\.(cia|3dsx|cetk|tik)$/i.test(file.name)) {
    notie.alert({
      type: 'error',
      text: `File name must be ends with .cia, .3dsx, .cetk, .tik<br>Your file name: ${file.name}`,
    })
    return;
  }

  console.log(file);
  config.file = file;
}

const onCancel = e => {
  config.file = null;
  config.progress = 0;
}

const onUpload = async e => {
  try {
    await axios.request({
      url: "/api/upload",
      method: "POST",
      params: {
        name: config.file.name,
      },
      data: config.file,
      onUploadProgress: ({loaded, total, progress, bytes, estimated, rate, upload = true}) => {
        console.log(loaded, total, progress);
        config.progress = progress * 100;
      },
    })
    await loadData();
    config.file = null;
    notie.alert({
      type: 'success',
      text: 'upload file successful',
    })
  } catch (error) {
    console.error(error);
    notie.alert({
      type: 'error',
      text: error.response.data['message'],
    })
  }

  config.progress = 0;
}

const deleteFile = async filename => {
  try {
    await axios.request({
      url: "/api/delete",
      method: "DELETE",
      params: {
        name: filename
      }
    })
    await loadData();
  } catch (error) {
    console.error(error)
    notie.alert({
      type: 'error',
      text: error.response.data['message'],
    })
  }
}

const sendGame = async name => {
  try {
    await axios.request({
      url: "/api/send",
      method: "POST",
      data: {
        address: config.address,
        name
      },
    })
    notie.alert({
      type: 'info',
      text: 'success to send the game URL to 3DS',
    });
  } catch (error) {
    console.error(error)
    notie.alert({
      type: 'error',
      text: error.response.data['message'],
    })
  }
}

const formatBytes = (bytes) => {
  const k = 1024,
    dm = 2,
    sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
    i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

const loadData = async () => {
  try {
    const response = await axios.get("/api/list", {params: {s: config.searchKeyword}});
    config.games = response.data;
  } catch (error) {
    console.error(error);
    notie.alert({
      type: 'error',
      text: error.response.data['message'],
    });
  }
}

onMounted(loadData);
onMounted(() => {
  config.address = window.localStorage.getItem("address") || "";
})
</script>

<style scoped lang="scss">
@mixin box() {
  border: 2px dashed #eee;
  height: 150px;
  cursor: pointer;
  position: relative;

  &:hover {
    border: 2px solid #2e94c4;

    .title {
      color: #1975A0;
    }
  }
}

.fileupload {
  @include box();

  .info {
    color: #A8A8A8;
    position: absolute;
    top: 50%;
    width: 100%;
    transform: translate(0, -50%);
    text-align: center;
  }

  .title {
    color: #787878;
  }

  input {
    position: absolute;
    cursor: pointer;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
  }

  .upload-limit-info {
    display: flex;
    justify-content: flex-start;
    flex-direction: column;
  }

  &.over {
    background: #5C5C5C;
    opacity: 0.8;
  }
}

.fileupload-progress {
  @include box();
}

table.table tr td {
  vertical-align: middle;
}
</style>
