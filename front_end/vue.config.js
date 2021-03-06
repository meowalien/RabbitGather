module.exports = {
  devServer: {
    sockHost: "peerjs.localhost",
    disableHostCheck: true,
    // host: "meowalien.com",
    // https: true,
    // public: "https://meowalien.com:8080/",
    // key: fs.readFileSync("meowalien.com.key"),
    // cert: fs.readFileSync("meowalien_com.pem"),
    proxy: {
      //設定代理
      "/api": {
        target: "https://api.localhost/", // 介面的域名
        changeOrigin: true,
        ws: true,
        pathRewrite: {
          "^/api": "", //萬用字元
        },
      },
    },
  },

  pwa: {
    name: "RabbitGather",
    themeColor: "#ffabab",
  },
};
