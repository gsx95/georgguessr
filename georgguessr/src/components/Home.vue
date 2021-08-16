<template>
  <el-dialog
    title="Tips"
    v-model="dialogVisible"
    width="30%"
    :before-close="handleClose"
  >
    <span>{{ $t("message.greeting", {}, { locale: "de" }) }}</span>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="dialogVisible = false"
          >Confirm</el-button
        >
      </span>
    </template>
  </el-dialog>
</template>

<script>
export default {
  name: "Home",
  data() {
    return {
      dialogVisible: true,
    };
  },
  methods: {
    handleClose(done) {
      this.$confirm("Are you sure to close this dialog?")
        .then(() => {
          done();
          this.dialogVisible = false;
        })
        .catch(() => {});
    },
  },
};
</script>

<style scoped>
</style>
