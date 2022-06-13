package middleware

import "testing"

func TestParser(t *testing.T) {
	r := HTMLParse(`


<!DOCTYPE html>
<html class="" lang="en">
<head prefix="og: http://ogp.me/ns#">
<meta charset="utf-8">
<link as="style" href="https://gitlab.bj.sensetime.com/assets/application-2cb8d6d6d17f1b1b8492581de92356755b864cbb6e48347a65baa2771a10ae4f.css" rel="preload">
<link as="style" href="https://gitlab.bj.sensetime.com/assets/highlight/themes/white-23255bab077a3cc8d17b4b7004aa866e340b0009c7897b509b5d0086710b698f.css" rel="preload">

<meta content="IE=edge" http-equiv="X-UA-Compatible">

<meta content="object" property="og:type">
<meta content="GitLab" property="og:site_name">
<meta content="Projects · Dashboard" property="og:title">
<meta content="GitLab SenseTime in Beijing" property="og:description">
<meta content="https://gitlab.bj.sensetime.com/assets/gitlab_logo-7ae504fe4f68fdebb3c2034e36621930cd36ea87924c11ff65dbcb8ed50dca58.png" property="og:image">
<meta content="64" property="og:image:width">
<meta content="64" property="og:image:height">
<meta content="https://gitlab.bj.sensetime.com/" property="og:url">
<meta content="summary" property="twitter:card">
<meta content="Projects · Dashboard" property="twitter:title">
<meta content="GitLab SenseTime in Beijing" property="twitter:description">
<meta content="https://gitlab.bj.sensetime.com/assets/gitlab_logo-7ae504fe4f68fdebb3c2034e36621930cd36ea87924c11ff65dbcb8ed50dca58.png" property="twitter:image">

<title>Projects · Dashboard · GitLab</title>
<meta content="Gi
`)
	if r.Error != nil {
		println(r.Error.Error())
		return
	}
	metas := r.FindAll("title")
	attrs := metas[0].Text()
	println(attrs)
}
