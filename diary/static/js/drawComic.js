full_screen.onclick=function(){
    if(full_screen.dataset['switch']==="off"){
        content.style.padding='0'
        content.style.fontSize='0'
        process.style.fontSize='medium'
        board.style.margin='0'
        board.style.width='100%'
        $('.content img').css('box-shadow','none')
        full_screen.dataset['switch']="on"
        full_screen.style.backgroundImage='url(\'/static/image/full-screen-off.png\')'
    }else{
        content.style.padding='0 12%'
        content.style.fontSize='medium'
        board.style.margin='50px auto'
        board.style.width='92%'
        $('.content img').css('box-shadow','0 0 16px rgb(0 0 0/50%)')
        full_screen.dataset['switch']="off"
        full_screen.style.backgroundImage='url(\'/static/image/full-screen-on.png\')'
    }
}
function onImageLoaded(e,t) {
    var a = t.getContext("2d")
        , n = e.width
        , d = e.naturalWidth
        , i = e.naturalHeight;
    t.width = d,
        t.height = i
    var r = e.dataset['index']
    for (var s = get_num(id, r), l = i % s, o = d, m = 0; m < s; m++) {
        var c = Math.floor(i / s)
            , g = c * m
            , h = i - c * (m + 1) - l;
        0 === m ? c += l : g += l,
            a.drawImage(e, 0, h, o, c, 0, g, o, c)
    }
    $(e).addClass("hide")
}
function get_num(e, t) {
    var a = 1;
    if (e>=220971){
        a=10
    }
    if (e >= 268850) {
        var n = e + t;
        switch (n = (n = (n = md5(n)).substr(-1)).charCodeAt(),
            n %= 10) {
            case 0:
                a = 2;
                break;
            case 1:
                a = 4;
                break;
            case 2:
                a = 6;
                break;
            case 3:
                a = 8;
                break;
            case 4:
                a = 10;
                break;
            case 5:
                a = 12;
                break;
            case 6:
                a = 14;
                break;
            case 7:
                a = 16;
                break;
            case 8:
                a = 18;
                break;
            case 9:
                a = 20
        }
    }
    return a
}
function loadComic(){
    delta+=5
    for(let i=0;i<step&&index<pageNum;i++,index++){
        let img= document.createElement('img')
        img.className='page'
        img.dataset['index']=index.toString().padStart(5,'0')
        img.style.backgroundImage='url(\'/static/image/loading.jpg\')'
        img.src="https://cdn-msp.18comic.org/media/photos/"+id+"/"+img.dataset['index']+".webp"
        img.onerror=function(){
            let src = this.src
            this.src='/static/image/loadfailed.png'
            this.onclick=function(){
                this.src=src
                this.onclick=null
            }
        }
        img.onload=function(){
            var t=document.createElement('canvas')
            this.after(t)
            onImageLoaded(this,t)
            delta--
        }
        images.appendChild(img)
    }
}
window.onscroll=function(){
    console.log(delta)
    if(index>=pageNum){
        loadEnd()
        window.onscroll=null
    }
    if(document.body.clientHeight-document.documentElement.scrollTop<2100&&delta<=10){
        loadComic()
        console.log('scroll loadComic')
    }
}
function loadStart(){
    process.querySelector('.text').innerText="本子正在加载中……"
    process.style.backgroundImage='url(\'/static/image/load.gif\')'
    process.style.backgroundPosition='center left -70%'
    process.style.backgroundSize='70%'
}
function loadEnd(){
    process.querySelector('.text').innerText="本子加载完毕！"
    process.style.backgroundImage='url(\'/static/image/load_success.png\')'
    process.style.backgroundPosition='center left 14%'
    process.style.backgroundSize='12%'
}
function loadFailed(){
    process.querySelector('.text').innerText="本子加载失败！"
    process.style.backgroundImage='url(\'/static/image/load_error.png\')'
    process.style.backgroundPosition='center left 14%'
    process.style.backgroundSize='12%'
}
loadStart()
if(pageNum===0){
    loadFailed()
}
loadComic()