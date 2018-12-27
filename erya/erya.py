import requests
from bs4 import BeautifulSoup
from selenium import webdriver
import requests
import re
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time
import os
def get_answer(title):
	url = "http://92daikan.com/tiku.aspx"
	headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"}
	data = {"__VIEWSTATE": "/wEPDwUKLTY1ODc5MTg4OA9kFgJmD2QWAgIDD2QWAgIBD2QWAgIDD2QWBAIBDw8WAh4HVmlzaWJsZWhkZAIDDw8WAh4EVGV4dAUJ5ZGo5bq3546LZGRkLELjXTA6DX1n5s0ZBfAzBWVh6LHCSFvuCwerteC9wIc=",
			'__VIEWSTATEGENERATOR':'C2F6633C',
'__EVENTVALIDATION': '/wEdAAMFpYViKQ4/8DOpRcWx3/h+9XDtxOeWDVhUDB3OayhOrqPpJucgJILabyLcPgnZw+nRvD/KT73i1xSUE/pjLhYQeVmmBedsdHf7buvq8aNAIQ==',
"ctl00$ContentPlaceHolder1$timu":  title,
'ctl00$ContentPlaceHolder1$gen': "查询"}
	time.sleep(1)
	req = requests.post(url=url,data=data,headers=headers).text
	soup = BeautifulSoup(req,"html.parser")
	answer = soup.find(id="daan").text
	return answer



def post():
	l = []
	browser = webdriver.Chrome()
	browser.get('http://passport2.chaoxing.com/login?fid=1719&refer=http://xiyou.fanya.chaoxing.com')
	i= 1
	while i:
		try:
			username = browser.find_element_by_class_name('zl_input')
			user = input("请输入学号")
			username.send_keys(str(user))
			password = browser.find_element_by_class_name('zl_input2')
			passs = input("请输入密码")
			password.send_keys(str(passs))
			yzm = browser.find_element_by_name('numcode')
			yzms = input("请输入验证码")
			yzm.send_keys(str(yzms))
			button = browser.find_element_by_class_name('zl_btn_right')
			button.click()
			browser.find_element_by_xpath('//a[@class="zaf_text"]').click()
			i = 0
		except Exception as e:
			print("请正确输入学号密码以及验证码")
			continue
	classs = input("请输入需要课程的地址：")
	browser.get(classs)
	button3 = browser.find_element_by_xpath("//a[contains(text(),'考试')]").click()
	button2 = browser.find_element_by_class_name('Btn_red_1').click()
	code = browser.find_element_by_name('identifyCodeRandom')
	codes = input('请输入验证码')
	code.send_keys(str(codes))
	button3 = browser.find_element_by_xpath('//a[@class="bluebtn"][@href="javascript:void(0)"][@onclick="startTest(\'1719\');"]').click()
	for i in range(1,100):
		try:
			time.sleep(2)
			a = browser.page_source
			wait = WebDriverWait(browser, 2)
			Id = re.findall('<span>当前第(.*?)题/共 .*? 题</span>',a,re.S)
			search = wait.until(EC.visibility_of_element_located((By.XPATH,"//div[@class='clearfix' and @style]")))
			answer = get_answer(search.text)
			if "#" in answer:
				lists = answer.split("#")
				for i in lists:
					time.sleep(1)
					a = "//a[contains(text(),'{0}')]".format(i)
					print(a)
					time.sleep(1)
					button5 = browser.find_element_by_xpath(a).click()			
				button4 = wait.until(EC.visibility_of_element_located((By.XPATH,"//a[contains(text(),'下一题')]"))).click()
			else:
				if answer == '错误':
					answer = "false"
				if answer =="正确":
					answer = "true"
				print(str(answer))
				if answer =="true":
					button7 = wait.until(EC.visibility_of_element_located((By.XPATH,'//b[@class="ri"]'))).click()
				elif answer =="false":
					button7 = wait.until(EC.visibility_of_element_located((By.XPATH,'//b[@class="wr"]'))).click()
				else:
					button5 = browser.find_element_by_link_text(str(answer)).click()
				#print(search.text)#获取题目
				button4 = wait.until(EC.visibility_of_element_located((By.XPATH,"//a[contains(text(),'下一题')]"))).click()
				time.sleep(1)
		except Exception as e:
			print(Id[0],'题')
			time.sleep(2)
			button4 = wait.until(EC.visibility_of_element_located((By.XPATH,"//a[contains(text(),'下一题')]"))).click()
			continue
	end = input("请查看是否全部答题完，按OK 结束")

post()
