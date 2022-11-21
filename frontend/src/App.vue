<template>
  <nav class="navbar bg-light">
    <div class="container">
      <a class="navbar-brand">FBI Remote Game Installer</a>
      <div class="d-flex">
        <button class="btn btn-primary me-2" type="button"
                @click="state.showConfig = !state.showConfig">Configuration
        </button>
        <div class="col-auto me-2">
          <input class="form-control" type="text" placeholder="Search Games" v-model="state.searchKeyword" @keydown.enter="loadData">
        </div>
        <button class="btn btn-success" type="button" @click="loadData">Search</button>
      </div>
    </div>
  </nav>
  <div class="container mt-4">
    <div v-if="!state.address" class="alert alert-warning" role="alert">
      Your 3DS address is empty, <a href="#" @click="state.showConfig = true" class="text-decoration-none">click
      here</a> to set up the address.
    </div>
    <div v-else class="alert alert-primary" role="alert">
      Your 3DS address is <strong>{{ state.address }}</strong>
    </div>

    <div class="mb-3" v-show="state.showConfig">
      <label for="exampleFormControlInput1" class="form-label">3DS Address</label>
      <input type="text" class="form-control" placeholder="" v-model="state.address">
    </div>

    <hr/>

    <div v-if="state.file == null" :class="['fileupload', state.dragging ? 'over': '']" @dragenter="state.dragging = true"
         @dragleave="state.dragging = false">
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
                 :style="{width: state.progress + '%'}">{{ state.file.name }}
            </div>
          </div>
          <div class="mt-2">
            File Name: {{ state.file.name }} File Size: {{ formatBytes(state.file.size) }}
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
      <tr v-for="game in state.games">
        <th scope="row">{{game.name}}</th>
        <td>{{ formatBytes(game.size) }}</td>
        <td>{{ dayjs.unix(game['mod_time']).format('YYYY-MM-DD HH:mm:ss') }}</td>
        <td class="text-end">
          <button type="button" class="btn btn-outline-danger btn-sm me-2" @click="deleteFile(game.name)">Delete</button>
          <button type="button" class="btn btn-outline-primary btn-sm"
                  :disabled="state.address === ''" @click="sendGame(game.name)">Send to 3DS</button>
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

const state = reactive({
  address: "",
  showConfig: false,
  file: null,
  games: [],
  progress: 0,
  searchKeyword: "",
  dragging: false
})

watch(() => state.address, (current, old) => {
  if (typeof current === "string") {
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
  state.file = file;
}

const onCancel = e => {
  state.file = null;
  state.progress = 0;
}

const onUpload = async e => {
  try {
    await axios.request({
      url: "/api/upload",
      method: "POST",
      params: {
        name: state.file.name,
      },
      data: state.file,
      onUploadProgress: ({loaded, total, progress, bytes, estimated, rate, upload = true}) => {
        console.log(loaded, total, progress);
        state.progress = progress * 100;
      },
    })
    await loadData();
    state.file = null;
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

  state.progress = 0;
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
        address: state.address,
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
    const response = await axios.get("/api/list", {params: {s: state.searchKeyword}});
    state.games = response.data;
  } catch (error) {
    console.error(error);
    notie.alert({
      type: 'error',
      text: error.response.data['message'] || 'something went wrong, please try again...',
    });
  }
}

onMounted(loadData);
onMounted(() => {
  state.address = window.localStorage.getItem("address") || "";
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
