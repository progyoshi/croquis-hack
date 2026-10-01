// アップロードボタンとファイル選択部分
const upload = document.getElementById("upload");
const filePlace =document.getElementById("filePlace");

const preview = document.getElementById("preview");

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

    if(!response.ok){
        console.log("error:", response.status);
        return;
    }

    const data = await response.json();

    const url = data.image_url;
    const conti_days = data.conti_days; // 継続日数
    const total_days = data.total_days; // 総日数

    console.log("url:"+url+" conti_days:"+conti_days+" total_days:"+total_days);

    preview.src = url;
});

window.onload = async() =>{
    const response = await fetch(
        "http://localhost:3000/api/today",
        {
            method: "GET"
        }
    );
    const data = await response.json();

    const url = data.image_url;
    const conti_days = data.conti_days; // 継続日数
    const total_days = data.total_days; // 総日数

    console.log("url:"+url+" conti_days:"+conti_days+" total_days:"+total_days);

    preview.src = url;
};
// preview.src = ここにpreviewのURLをいれる;