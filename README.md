# Подключение ТВ

когда создатель презентации подключается к сессии(вебсоккет), генерится `tv_token`
и посылается в ответе. На тв делается пост запрос с `tv_token` form field и с хедером `X-Engagers-tvOS-UUID = true`
в ответ получает урл нужной сессии `ws://engagers.herokuapp.com/ws/1/23b33cd8-3f43-4fd4-9c9e-95d5267d48e6`