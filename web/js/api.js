// アップロードボタンとファイル選択部分
const upload = document.getElementById("upload");
const filePlace =document.getElementById("filePlace");

// 変更したら入れようとか思ったけどボタンを押したらひっぱるでよかった子たち
// filePlace.addEventListener("change",function(){
//     const file = filePlace.files[0];
// });

upload.addEventListener("click",async () => {
    // 画像入れます
    const file = filePlace.files[0];
    // でーたのいれものつくります
    const formData = new FormData();
    // つくったいれものにがぞういれます
    formData.append("test", file)
    // goのapiにむけて画像はいったいれものをhttpつうしんでおくる
    const response = await fetch(
        "http://localhost:3000/api/today",
        {
            method: "POST",
            body: formData
        }
    );
});

const reload = document.getElementById("reload")
const preview = document.getElementById("preview");

reload.addEventListener("click",async()=>{
    const response = await fetch(
    "http://localhost:3000/api/today"
    );

    const blob = await response.blob();

    const url = URL.createObjectURL(blob);

    preview.src = url;
});


// preview.src = ここにpreviewのURLをいれる;