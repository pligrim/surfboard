# /usr/bin/env/ python

"""

Module Docstring:       This will create a dynamic confluence page for a Surfboard under the space: https://confluence.ipttools.info/display/SC/Status+Checking


Pre-requisites:         Beautiful soup - https://www.pythonforbeginners.com/beautifulsoup/beautifulsoup-4-python
                        PrettyTable -  http://zetcode.com/python/prettytable/

"""
# Internal Imports
import confluence_utilities


# External Imports
import os
import HTMLParser

# Error Reporting
import argparse
import logging


def update_page(value, user, passwd, page_id, ancestors_id, version, page_title, space_key):
    update_page_payload = confluence_utilities.update_page_payload(value, version, page_id, ancestors_id, page_title, space_key)
    response = confluence_utilities.update_page(update_page_payload, page_id, user, passwd)
    return response


def create_page(value, title, user, passwd, space_key, ancestors_id):
    create_page_payload = confluence_utilities.create_page_payload(value, ancestors_id, title, space_key)
    response = confluence_utilities.create_page(create_page_payload, user, passwd)
    return response

    

def get_files(current_dir):
    """
    This method will do an os.walk on the index directory (created by the jenkinfile called: 'Jenkinsfile.dashboard')
        and store all the index.html files into a list called files_list.

    :return:                This method will return the following variables:
                                files_list - list of all the index.html
    """

    files_list = []

    for root, dirs, files in os.walk(current_dir):
        for filename in files:
            files_list.append(os.path.join(root, filename))
    return files_list


def check_for_page(parsed_user, parsed_passwd, parsed_space_key, ITF_project_name, filename):
    """

    ...requires updating cassie...

    :param parsed_user:            Username for Confluence
    :param parsed_passwd:          Password for Confluence
    :param parsed_space_key:       SpaceKey for Confluence
    :return:                       This method will return the following variables:
                                          page_title, method (post or put), page_id and version.
    """

    try:
        # Variables
        fname = str.replace(str.replace(str.replace(filename,"-map.insert","",1),"-"," ",-1),"./"," ",1)
        title = "{ITF} - Surfboard Report - {FName}".format(ITF=ITF_project_name, FName=fname)
      
        is_exists, response = confluence_utilities.check_if_page_exists(title, parsed_space_key, parsed_user, parsed_passwd)

        if is_exists is False:
            # Create new page
            method = "post"
            # No page ID required.
            return title, method, None, 1

        elif is_exists is True:
            # Pages exist therefore put method, ID and version are required.
            page_title = response['results'][0]['title']
            page_id = response['results'][0]['id']
            version = response['results'][0]['version']['number']
            method = "put"

            # Checking..
            if version == "":
                # if a page has an empty version, it will be parse as 1 instead of None
                return page_title, method, page_id, 1
            else:
                return page_title, method, page_id, version

    except Exception as msg:
        print("The error message is:  {msg}".format(msg=msg))
        exit(1)


def handle_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument("space_key", type=str, help="Space key for confluence...")
    parser.add_argument("anchor_page_id", type=str, help="Page Id for the anchor page...") # https://confluence.ipttools.info/rest/api/content?type=page&title= `{YOUR ANCHOR PAGE TITLE}` -
    parser.add_argument("ITF_project_name", type=str, help="e.g EUE ...")
    return parser.parse_args()


if __name__ == '__main__':
    args = handle_arguments()

    # Env Variables:
    user = os.environ.get('CONF_USER')
    passwd = os.environ.get('CONF_PASSWORD')

    # Parsed Variables:
    space_key = args.space_key
    anchor_page_id = args.anchor_page_id
    ITF_project_name = args.ITF_project_name

    # Validation that envs are not None or anything else:
    invalid_env_vars = [None]
    if any(True for x in invalid_env_vars if x in [user, passwd]):
        exit("Username and password are None please set credentials")


    ltf_files = get_files('./')

    deployments_files = filter(lambda x: x.endswith('.insert'), ltf_files)

    deployments_files.sort()

    for deployments_file in deployments_files:

        with open(deployments_file, 'r') as myfile:
            data = myfile.read()

        # Method calls:
        title, method, page_id, version = check_for_page(user, passwd, space_key, ITF_project_name, deployments_file)

        # Checks if a new or old page:
        dashboard = HTMLParser.HTMLParser().unescape(data)

        if method == "put":
            print("put:")
            response = update_page(dashboard, user, passwd, page_id, anchor_page_id, version, title, space_key)
            # Will fail or pass the jenkins build
            #output = "Success" if int(response) is 200 else "Failed my response code is {r}".format(r=response)
            print(response)

        if method == "post":
            print("post:")
            response = create_page(dashboard, title, user, passwd, space_key, anchor_page_id)
            # Will fail or pass the jenkins build
            #output = "Success" if int(response) is 200 else "Failed my response code is {r}".format(r=response)
            print(response)

