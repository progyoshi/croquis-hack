const preview = document.getElementById("preview");
const upload = document.getElementById("upload");
const filePlace =document.getElementById("filePlace");

filePlace.addEventListener("change",function(){
    const file = filePlace.files[0];
});

const formData = new FormData();

formData.append("image", file)

upload.addEventListener("click",function(){

});

preview.src = ここにpreviewのURLをいれる;