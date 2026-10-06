<template>
    <section class="welcome-section">

        <div class="welcome-content">

            <div class="logo">N</div>

            <p class="eyebrow">
                WELCOME BACK
            </p>

            <h1>
                Your people.<br>
                Your <span>place.</span>
            </h1>

            <p class="welcome-text">
                Connect with people, share your thoughts,
                and keep up with the things that matter to you.
            </p>

            <div class="welcome-line"></div>

            <p class="small-text">
                A simple place to stay connected.
            </p>

        </div>

        <div class="decoration decoration-one"></div>
        <div class="decoration decoration-two"></div>
        <div class="decoration decoration-three"></div>

        <div class="circle circle-one"></div>
        <div class="circle circle-two"></div>

        <div class="grid-pattern"></div>

    </section>
</template>

<style scoped>
.welcome-section {
    position: relative;

    min-height: 100vh;
    min-height: 100dvh;

    display: flex;
    align-items: center;

    padding: 80px;

    background: var(--bg-color);

    border-right: 2px solid var(--main-color);

    overflow: hidden;
}

.welcome-content {
    position: relative;
    z-index: 5;

    max-width: 650px;
}

.logo {
    width: 55px;
    height: 55px;

    display: flex;
    align-items: center;
    justify-content: center;

    margin-bottom: 55px;

    background: var(--input-focus);

    border: 2px solid var(--main-color);
    border-radius: 8px;

    box-shadow: 5px 5px var(--main-color);

    color: white;

    font-family: "JetBrains Mono", monospace;
    font-size: 25px;
    font-weight: 600;
}

.eyebrow {
    margin: 0 0 15px;

    color: var(--input-focus);

    font-family: "JetBrains Mono", monospace;

    font-size: 13px;
    font-weight: 600;

    letter-spacing: 3px;
}

h1 {
    margin: 0;

    color: var(--main-color);

    font-family: "Liter", serif;

    font-size: clamp(60px, 6vw, 100px);

    line-height: 0.95;

    letter-spacing: -4px;
}

h1 span {
    color: var(--input-focus);
}

.welcome-text {
    max-width: 480px;

    margin: 35px 0 30px;

    color: var(--font-color-sub);

    font-size: 18px;

    line-height: 1.7;
}

.welcome-line {
    width: 70px;
    height: 4px;

    margin-bottom: 20px;

    background: var(--main-color);
}

.small-text {
    margin: 0;

    color: var(--main-color);

    font-family: "JetBrains Mono", monospace;

    font-size: 12px;
}

.decoration {
    position: absolute;

    border: 2px solid var(--main-color);

    background: var(--input-focus);

    box-shadow: 7px 7px var(--main-color);
}

.decoration-one {
    width: 180px;
    height: 180px;

    right: -70px;
    top: 15%;

    transform: rotate(25deg);
}

.decoration-two {
    width: 100px;
    height: 100px;

    right: 18%;
    bottom: -35px;

    background: var(--bg-color);

    transform: rotate(15deg);

    box-shadow: 5px 5px var(--main-color);
}

.decoration-three {
    width: 50px;
    height: 50px;

    left: 8%;
    bottom: 12%;

    background: var(--main-color);

    transform: rotate(45deg);

    box-shadow: none;
}

.circle {
    position: absolute;

    border: 2px solid var(--main-color);

    border-radius: 50%;
}

.circle-one {
    width: 230px;
    height: 230px;

    right: 10%;
    top: -90px;
}

.circle-two {
    width: 90px;
    height: 90px;

    left: 40%;
    bottom: -45px;

    background: var(--input-focus);
}

.grid-pattern {
    position: absolute;

    right: 0;
    bottom: 0;

    width: 320px;
    height: 320px;

    opacity: 0.12;

    background-image:
        linear-gradient(
            var(--main-color) 1px,
            transparent 1px
        ),
        linear-gradient(
            90deg,
            var(--main-color) 1px,
            transparent 1px
        );

    background-size: 25px 25px;

    mask-image: linear-gradient(
        to top left,
        black,
        transparent
    );
}

@media (max-width: 900px) {
    .welcome-section {
        min-height: 430px;

        padding: 50px;

        border-right: none;

        border-bottom: 2px solid var(--main-color);
    }

    h1 {
        font-size: clamp(50px, 10vw, 80px);
    }

    .decoration-one {
        right: -90px;
    }

    .circle-one {
        right: -70px;
    }
}

@media (max-width: 500px) {
    .welcome-section {
        min-height: 390px;

        padding: 35px 25px;
    }

    .logo {
        width: 50px;
        height: 50px;

        margin-bottom: 35px;
    }

    h1 {
        font-size: 50px;

        letter-spacing: -2px;
    }

    .welcome-text {
        margin-top: 25px;

        font-size: 15px;
    }

    .decoration-one {
        right: -120px;
    }

    .circle-one {
        right: -100px;
    }
}

@media (max-width: 640px) {
    .welcome-section {
        display: none;
    }
}
</style>
