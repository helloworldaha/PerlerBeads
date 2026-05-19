const app = getApp()

Page({
  data: {
    canvasWidth: 0,
    canvasHeight: 0,
    grid: [],
    gridWidth: 0,
    gridHeight: 0,
    palette: [],
    colorCounts: {},
    totalBeads: 0,
    scale: 1.0,
    showGrid: true,
    exporting: false,
    patternId: '',
    touchStartDistance: 0,
    touchStartScale: 1.0
  },

  onLoad(options) {
    if (options.data) {
      try {
        const data = JSON.parse(decodeURIComponent(options.data))
        this.initPage(data)
      } catch (err) {
        console.error('解析数据失败', err)
        wx.showToast({
          title: '数据解析失败',
          icon: 'none',
          duration: 2000
        })
      }
    }
  },

  onReady() {
    this.initCanvas()
  },

  initPage(data) {
    const { grid, palette, id } = data
    const gridWidth = grid[0]?.length || 0
    const gridHeight = grid.length

    const colorCounts = {}
    let totalBeads = 0

    grid.forEach(row => {
      row.forEach(colorIndex => {
        colorCounts[colorIndex] = (colorCounts[colorIndex] || 0) + 1
        totalBeads++
      })
    })

    const paletteWithCount = palette.map((color, index) => ({
      ...color,
      count: colorCounts[index] || 0
    })).sort((a, b) => b.count - a.count)

    this.setData({
      grid,
      gridWidth,
      gridHeight,
      palette: paletteWithCount,
      colorCounts,
      totalBeads,
      patternId: id
    })
  },

  initCanvas() {
    const query = wx.createSelectorQuery()
    query.select('.canvas-container')
      .fields({ size: true })
      .exec((res) => {
        if (!res[0]) return

        const containerWidth = res[0].width
        const containerHeight = res[0].height
        const maxSize = Math.min(containerWidth, containerHeight) - 20
        const gridLineWidth = 1

        const cellSize = Math.floor((maxSize - (Math.max(this.data.gridWidth, this.data.gridHeight) + 1) * gridLineWidth) / Math.max(this.data.gridWidth, this.data.gridHeight))
        const canvasWidth = this.data.gridWidth * cellSize + (this.data.gridWidth + 1) * gridLineWidth
        const canvasHeight = this.data.gridHeight * cellSize + (this.data.gridHeight + 1) * gridLineWidth

        const canvasQuery = wx.createSelectorQuery()
        canvasQuery.select('#gridCanvas')
          .fields({ node: true, size: true })
          .exec((canvasRes) => {
            if (!canvasRes[0]) return

            const canvas = canvasRes[0].node
            const ctx = canvas.getContext('2d')
            const dpr = wx.getSystemInfoSync().pixelRatio

            canvas.width = canvasWidth * dpr
            canvas.height = canvasHeight * dpr
            ctx.scale(dpr, dpr)

            this.setData({
              canvasWidth,
              canvasHeight,
              _canvas: canvas,
              _ctx: ctx,
              _cellSize: cellSize,
              _gridLineWidth: gridLineWidth
            })

            this.renderCanvas()
          })
      })
  },

  renderCanvas() {
    const { _canvas, _ctx, grid, gridWidth, gridHeight, _cellSize, _gridLineWidth, scale, showGrid } = this.data
    if (!_canvas || !_ctx || !grid.length) return

    _ctx.clearRect(0, 0, _canvas.width, _canvas.height)

    _ctx.save()
    _ctx.scale(scale, scale)

    const gridColor = 'rgba(200, 200, 200, 1)'
    
    if (showGrid) {
      _ctx.fillStyle = gridColor
      const fullWidth = gridWidth * _cellSize + (gridWidth + 1) * _gridLineWidth
      const fullHeight = gridHeight * _cellSize + (gridHeight + 1) * _gridLineWidth
      _ctx.fillRect(0, 0, fullWidth, fullHeight)

      for (let y = 0; y < gridHeight; y++) {
        for (let x = 0; x < gridWidth; x++) {
          const colorIndex = grid[y][x]
          const color = this.data.palette[colorIndex]

          if (color) {
            _ctx.fillStyle = color.hex
            const startX = x * (_cellSize + _gridLineWidth) + _gridLineWidth
            const startY = y * (_cellSize + _gridLineWidth) + _gridLineWidth
            _ctx.fillRect(startX, startY, _cellSize, _cellSize)
          }
        }
      }
    } else {
      for (let y = 0; y < gridHeight; y++) {
        for (let x = 0; x < gridWidth; x++) {
          const colorIndex = grid[y][x]
          const color = this.data.palette[colorIndex]

          if (color) {
            _ctx.fillStyle = color.hex
            _ctx.fillRect(x * _cellSize, y * _cellSize, _cellSize, _cellSize)
          }
        }
      }
    }

    _ctx.restore()
  },

  toggleGrid() {
    this.setData({
      showGrid: !this.data.showGrid
    }, () => {
      this.renderCanvas()
    })
  },

  zoomIn() {
    if (this.data.scale < 3) {
      this.setData({
        scale: Math.min(3, this.data.scale + 0.2)
      }, () => {
        this.renderCanvas()
      })
    }
  },

  zoomOut() {
    if (this.data.scale > 0.3) {
      this.setData({
        scale: Math.max(0.3, this.data.scale - 0.2)
      }, () => {
        this.renderCanvas()
      })
    }
  },

  getDistance(touches) {
    const dx = touches[0].clientX - touches[1].clientX
    const dy = touches[0].clientY - touches[1].clientY
    return Math.sqrt(dx * dx + dy * dy)
  },

  handleTouchStart(e) {
    if (e.touches.length === 2) {
      this.setData({
        touchStartDistance: this.getDistance(e.touches),
        touchStartScale: this.data.scale
      })
    }
  },

  handleTouchMove(e) {
    if (e.touches.length === 2) {
      const currentDistance = this.getDistance(e.touches)
      const scaleRatio = currentDistance / this.data.touchStartDistance
      let newScale = this.data.touchStartScale * scaleRatio

      newScale = Math.max(0.3, Math.min(3, newScale))

      if (Math.abs(newScale - this.data.scale) > 0.01) {
        this.setData({ scale: newScale }, () => {
          this.renderCanvas()
        })
      }
    }
  },

  handleTouchEnd() {
    this.setData({
      touchStartDistance: 0,
      touchStartScale: 1.0
    })
  },

  selectColor(e) {
    const index = e.currentTarget.dataset.index
    console.log('选择颜色', this.data.palette[index])
  },

  goBack() {
    wx.navigateBack()
  },

  exportImage() {
    if (!this.data.patternId) {
      wx.showToast({
        title: '无法导出',
        icon: 'none',
        duration: 2000
      })
      return
    }

    this.setData({ exporting: true })

    const exportUrl = `${app.globalData.apiBaseUrl}/export?id=${this.data.patternId}`
    console.log('导出URL:', exportUrl)

    wx.downloadFile({
      url: exportUrl,
      success: (res) => {
        console.log('下载响应:', res)
        if (res.statusCode === 200) {
          console.log('临时文件路径:', res.tempFilePath)
          
          if (!res.tempFilePath) {
            wx.showToast({
              title: '下载失败：临时文件路径为空',
              icon: 'none',
              duration: 2000
            })
            return
          }

          wx.saveImageToPhotosAlbum({
            filePath: res.tempFilePath,
            success: () => {
              wx.showToast({
                title: '保存成功',
                icon: 'success',
                duration: 2000
              })
            },
            fail: (err) => {
              console.log('保存图片失败详情:', JSON.stringify(err))
              if (err.errMsg.includes('auth') || err.errMsg.includes('denied')) {
                wx.showModal({
                  title: '提示',
                  content: '需要授权保存图片到相册，请在设置中开启相册权限',
                  success: (res) => {
                    if (res.confirm) {
                      wx.openSetting()
                    }
                  }
                })
              } else if (err.errMsg.includes('no such file or directory')) {
                wx.showToast({
                  title: '临时文件不存在，请重试',
                  icon: 'none',
                  duration: 2500
                })
              } else {
                wx.showToast({
                  title: '保存失败：' + (err.errMsg || '未知错误'),
                  icon: 'none',
                  duration: 2500
                })
              }
            }
          })
        } else {
          console.log('下载失败，状态码:', res.statusCode)
          wx.showToast({
            title: '下载失败：状态码 ' + res.statusCode,
            icon: 'none',
            duration: 2000
          })
        }
      },
      fail: (err) => {
        console.log('下载图片失败详情:', JSON.stringify(err))
        wx.showToast({
          title: '下载失败，请检查网络连接',
          icon: 'none',
          duration: 2000
        })
      },
      complete: () => {
        this.setData({ exporting: false })
      }
    })
  }
})