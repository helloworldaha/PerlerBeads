const app = getApp()

Page({
  data: {
    selectedImage: '',
    size: 32,
    colorLimit: 16,
    forceCrop: false,
    loading: false
  },

  onLoad() {
    console.log('首页加载完成')
  },

  chooseImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath
        const fileSize = res.tempFiles[0].size

        if (fileSize > 2 * 1024 * 1024) {
          wx.showToast({
            title: '图片大小不能超过 2MB',
            icon: 'none',
            duration: 2000
          })
          return
        }

        this.setData({
          selectedImage: tempFilePath
        })
      },
      fail: (err) => {
        console.log('选择图片失败', err)
      }
    })
  },

  selectSize(e) {
    const size = parseInt(e.currentTarget.dataset.size)
    this.setData({ size })
  },

  selectColorLimit(e) {
    const colorLimit = parseInt(e.currentTarget.dataset.limit)
    this.setData({ colorLimit })
  },

  selectScaleMode(e) {
    const forceCrop = e.currentTarget.dataset.crop === 'true'
    this.setData({ forceCrop })
  },

  generatePattern() {
    if (!this.data.selectedImage) {
      wx.showToast({
        title: '请先选择图片',
        icon: 'none',
        duration: 2000
      })
      return
    }

    this.setData({ loading: true })

    const uploadTask = wx.uploadFile({
      url: `${app.globalData.apiBaseUrl}/generate`,
      filePath: this.data.selectedImage,
      name: 'image',
      formData: {
        size: this.data.size.toString(),
        color_limit: this.data.colorLimit.toString(),
        force_crop: this.data.forceCrop.toString()
      },
      success: (res) => {
        this.setData({ loading: false })

        if (res.statusCode === 200) {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            wx.navigateTo({
              url: `/pages/result/result?data=${encodeURIComponent(JSON.stringify(data.data))}`
            })
          } else {
            wx.showToast({
              title: data.message || '生成失败',
              icon: 'none',
              duration: 2000
            })
          }
        } else {
          wx.showToast({
            title: '服务器错误',
            icon: 'none',
            duration: 2000
          })
        }
      },
      fail: (err) => {
        this.setData({ loading: false })
        console.log('上传失败', err)
        wx.showToast({
          title: '上传失败，请检查网络',
          icon: 'none',
          duration: 2000
        })
      }
    })

    uploadTask.onProgressUpdate((res) => {
      console.log('上传进度', res.progress)
      console.log('已上传', res.totalBytesSent, '总大小', res.totalBytesExpectedToSend)
    })
  }
})